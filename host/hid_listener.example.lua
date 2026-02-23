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
end
