package handler

import (
	"seanAIgent/internal/service"
	"sync"

	"github.com/94peter/vulpes/ezapi"
	"github.com/gin-gonic/gin"
)

type notificationAPI struct {
	svc service.Service
}

var initNotificationApiOnce sync.Once

func InitNotificationApi(service service.Service) {
	initBookingApiOnce.Do(func() {
		api := &notificationAPI{
			svc: service,
		}

		ezapi.RegisterGinApi(func(r ezapi.RouterGroup) {
			// 推播學生出席狀況
			r.GET("/notification/attendance", api.attendance)
		})
	})
}

func (n *notificationAPI) attendance(c *gin.Context) {
	c.JSON(200, data)
}

var data = `
{
  "type": "bubble",
  "size": "mega",
  "body": {
    "type": "box",
    "layout": "vertical",
    "spacing": "md",
    "contents": [
      {
        "type": "text",
        "text": "🏊‍♂️ 訓練出席狀況通知",
        "weight": "bold",
        "size": "xl",
        "color": "#0B62E3"
      },
      {
        "type": "text",
        "text": "2025/01/01 - 2025/01/07",
        "size": "sm",
        "color": "#888888",
        "margin": "sm"
      },
      {
        "type": "separator",
        "margin": "md"
      },
      {
        "type": "box",
        "layout": "vertical",
        "margin": "md",
        "contents": [
          {
            "type": "text",
            "text": "👦 學生姓名：王小明",
            "size": "md",
            "weight": "bold"
          },
          {
            "type": "box",
            "layout": "vertical",
            "margin": "md",
            "spacing": "xs",
            "contents": [
              {
                "type": "text",
                "text": "📅 預約課程",
                "weight": "bold",
                "color": "#555555"
              },
              {
                "type": "text",
                "text": "1/2、1/4、1/6",
                "wrap": true,
                "margin": "xs"
              }
            ]
          },
          {
            "type": "box",
            "layout": "vertical",
            "margin": "md",
            "spacing": "xs",
            "contents": [
              {
                "type": "text",
                "text": "🟢 出席紀錄",
                "weight": "bold",
                "color": "#1E9E3A"
              },
              {
                "type": "text",
                "text": "1/2、1/6",
                "wrap": true,
                "margin": "xs"
              }
            ]
          },
          {
            "type": "box",
            "layout": "vertical",
            "margin": "md",
            "spacing": "xs",
            "contents": [
              {
                "type": "text",
                "text": "🔴 缺席紀錄",
                "weight": "bold",
                "color": "#D23339"
              },
              {
                "type": "text",
                "text": "1/4（未請假）",
                "wrap": true,
                "margin": "xs"
              }
            ]
          },
          {
            "type": "separator",
            "margin": "md"
          },
          {
            "type": "box",
            "layout": "horizontal",
            "margin": "md",
            "contents": [
              {
                "type": "box",
                "layout": "vertical",
                "contents": [
                  {
                    "type": "text",
                    "text": "總課程",
                    "color": "#555555",
                    "size": "sm"
                  },
                  {
                    "type": "text",
                    "text": "3",
                    "weight": "bold",
                    "size": "lg"
                  }
                ]
              },
              {
                "type": "box",
                "layout": "vertical",
                "contents": [
                  {
                    "type": "text",
                    "text": "出席",
                    "color": "#1E9E3A",
                    "size": "sm"
                  },
                  {
                    "type": "text",
                    "text": "2",
                    "weight": "bold",
                    "size": "lg"
                  }
                ]
              },
              {
                "type": "box",
                "layout": "vertical",
                "contents": [
                  {
                    "type": "text",
                    "text": "缺席",
                    "color": "#D23339",
                    "size": "sm"
                  },
                  {
                    "type": "text",
                    "text": "1",
                    "weight": "bold",
                    "size": "lg"
                  }
                ]
              }
            ]
          },
          {
            "type": "text",
            "text": "📈 出席率：67%",
            "size": "md",
            "weight": "bold",
            "margin": "md",
            "color": "#0B62E3"
          }
        ]
      }
    ]
  }
}
`
