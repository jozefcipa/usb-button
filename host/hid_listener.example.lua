-- Move this to ~/hid_listener.lua
-- --------------------------------

local log_file = "/tmp/usb_button.log"

local function write_event(label)
  os.execute(string.format("echo '%s' >> %s", label, log_file))
end

function onSinglePress()
  write_event("short")
end

function onDoublePress()
  write_event("double")
end

function onLongPress()
  write_event("long")
  -- led_on()  -- turn Pico LED on
  -- led_off() -- turn Pico LED off
  -- led_blink() -- put Pico LED back in blink mode
end
