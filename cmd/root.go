/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/94peter/vulpes/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "seanAIgent",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.seanAIgent.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in home directory with name ".seanAIgent" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".seanAIgent")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	log.SetConfig(
		log.WithDev(viper.GetBool("log.dev")),
		log.WithLevel(viper.GetString("log.level")),
		log.WithCallerSkip(1),
		log.WithServiceName("seanAIgent"),
		log.WithEnv(viper.GetString("log.env")),
	)
}

func initTracer(ctx context.Context) (func(ctx context.Context), error) {
	service := viper.GetString("tracing.service")
	if service == "" {
		return nil, fmt.Errorf("tracing.service is required")
	}
	endpoint := viper.GetString("tracing.endpoint")
	log.Info("tracing init:",
		log.String("endpoint", endpoint),
		log.String("env", viper.GetString("tracing.env")),
	)
	t, err := newOtelTracer(ctx, service)
	if err != nil {
		return nil, err
	}
	return t.Start()
}

func newOtelTracer(ctx context.Context, service string) (*otelTracer, error) {
	endpoint := viper.GetString("tracing.endpoint")
	sample := viper.GetFloat64("tracing.sample")
	log.Info("tracing init:",
		log.String("endpoint", endpoint),
		log.Float64("sample", sample),
	)
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	return &otelTracer{
		exporter:       exporter,
		resource:       getResource(service),
		sample:         getSampler(sample),
		metricExporter: metricExporter,
	}, nil
}

func getResource(service string) *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(service), // 服務名稱
	)
}

func getSampler(sample float64) sdktrace.Sampler {
	if sample <= 0 {
		return sdktrace.ParentBased(sdktrace.NeverSample())
	}
	if sample == 1 {
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sample))
}

const defaultMetricInterval = 30 * time.Second

type otelTracer struct {
	exporter       sdktrace.SpanExporter
	metricExporter metric.Exporter
	resource       *resource.Resource
	sample         sdktrace.Sampler
}

func (t *otelTracer) Start() (func(ctx context.Context), error) {
	// Start a new tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(t.sample),
		sdktrace.WithBatcher(t.exporter),
		sdktrace.WithResource(t.resource),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Start a new matrics provider
	mp := metric.NewMeterProvider(
		metric.WithResource(t.resource),
		metric.WithReader(metric.NewPeriodicReader(t.metricExporter,
			metric.WithInterval(defaultMetricInterval))),
	)

	// 設定全域 MeterProvider
	otel.SetMeterProvider(mp)
	err := runtime.Start(
		runtime.WithMinimumReadMemStatsInterval(defaultMetricInterval),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start runtime instrumentation: %w", err)
	}

	return func(ctx context.Context) {

		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down tracer provider: %v", err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			log.Fatalf("Error shutting down meter provider: %v", err)
		}

	}, nil
}
