(function () {
  'use strict';
  
  // Utility functions
  function parseISODate(isoString) {
    if (!isoString) return null;
    const parts = isoString.match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (!parts) return null;
    
    const year = parseInt(parts[1], 10);
    const month = parseInt(parts[2], 10) - 1;
    const day = parseInt(parts[3], 10);
    const date = new Date(Date.UTC(year, month, day));
    
    if (date.getUTCFullYear() === year && 
        date.getUTCMonth() === month && 
        date.getUTCDate() === day) {
      return date;
    }
    return null;
  }
  
  function formatDate(date, format, locale) {
    if (!date || isNaN(date.getTime())) return "";
    
    const options = { timeZone: "UTC" };
    const formatMap = {
      "locale-short": "short",
      "locale-long": "long",
      "locale-full": "full",
      "locale-medium": "medium"
    };
    
    options.dateStyle = formatMap[format] || "medium";
    
    try {
      return new Intl.DateTimeFormat(locale, options).format(date);
    } catch (e) {
      // Fallback to ISO format
      const year = date.getUTCFullYear();
      const month = (date.getUTCMonth() + 1).toString().padStart(2, "0");
      const day = date.getUTCDate().toString().padStart(2, "0");
      return `${year}-${month}-${day}`;
    }
  }
  
  // Find elements
  function findElements(trigger) {
    const calendarId = trigger.id + "-calendar-instance";
    const calendar = document.getElementById(calendarId);
    const hiddenInput = document.getElementById(trigger.id + "-hidden") || 
                       trigger.parentElement?.querySelector("[data-tui-datepicker-hidden-input]");
    const display = trigger.querySelector("[data-tui-datepicker-display]");
    
    return { calendar, hiddenInput, display };
  }
  
  // Update display
  function updateDisplay(trigger) {
    const elements = findElements(trigger);
    if (!elements.display || !elements.hiddenInput) return;
    
    const format = trigger.getAttribute("data-tui-datepicker-display-format") || "locale-medium";
    const locale = trigger.getAttribute("data-tui-datepicker-locale-tag") || "en-US";
    const placeholder = trigger.getAttribute("data-tui-datepicker-placeholder") || "Select a date";
    
    if (elements.hiddenInput.value) {
      const dateStrings = elements.hiddenInput.value.split(',').filter(d => d);
      const dates = dateStrings.map(s => parseISODate(s)).filter(d => d);

      if (dates.length > 0) {
        let displayText;
        if (dates.length === 1) {
          displayText = formatDate(dates[0], format, locale);
        } else if (dates.length === 2) {
          displayText = dates.map(d => formatDate(d, format, locale)).join(', ');
        } else {
          displayText = `${dates.length} dates selected`;
        }
        
        elements.display.textContent = displayText;
        elements.display.classList.remove("text-muted-foreground");
        return;
      }
    }
    
    elements.display.textContent = placeholder;
    elements.display.classList.add("text-muted-foreground");
  }
  
  // Handle calendar date selection
  document.addEventListener("calendar-date-selected", (e) => {
    // Find the datepicker trigger associated with this calendar
    const calendar = e.target;
    if (!calendar || !calendar.id.endsWith("-calendar-instance")) return;
    
    const triggerId = calendar.id.replace("-calendar-instance", "");
    const trigger = document.getElementById(triggerId);
    if (!trigger || !trigger.hasAttribute("data-tui-datepicker")) return;

    const multiple = trigger.getAttribute("data-tui-datepicker-multiple") === "true";
    console.log("Popover check -> Is multiple:", multiple, "Attribute value:", trigger.getAttribute("data-tui-datepicker-multiple"));

    // The hidden input is already updated by calendar.js.
    // Just tell the datepicker to update its display from the hidden input.
    updateDisplay(trigger);
    
    // Close the popover ONLY in single-select mode
    if (!multiple && window.closePopover) {
      const popoverId = trigger.getAttribute("aria-controls") || (trigger.id + "-content");
      window.closePopover(popoverId);
    }
  });
  
  // Form reset handling
  document.addEventListener("reset", (e) => {
    if (!e.target.matches("form")) return;
    
    e.target.querySelectorAll('[data-tui-datepicker="true"]').forEach(trigger => {
      const elements = findElements(trigger);
      if (elements.hiddenInput) {
        elements.hiddenInput.value = "";
      }
      updateDisplay(trigger);
    });
  });
  
  // MutationObserver for initial display update
  new MutationObserver(() => {
    document.querySelectorAll('[data-tui-datepicker="true"]:not([data-rendered])').forEach(trigger => {
      trigger.setAttribute('data-rendered', 'true');
      updateDisplay(trigger);
    });
  }).observe(document.body, { childList: true, subtree: true });
})();