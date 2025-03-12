module("luci.controller.8311", package.seeall)

local util = require "luci.util"
local i18n = require "luci.i18n"
local fs = require "nixio.fs"
local bit = require "nixio.bit"
local uci = require "luci.model.uci"

function system_status()
	local hostname = util.exec("uname -n"):trim()
	local cpuinfo = util.exec("cat /proc/cpuinfo "):trim()
	local machine = cpuinfo:match("machine%s+:%s+(%S+)")
	local architecture = cpuinfo:match("system type%s+:%s+(%S+)")
	local kernel_release = util.exec("uname -r"):trim()
	local openwrt_release = util.exec("source /etc/openwrt_release; echo $DISTRIB_DESCRIPTION"):trim()
	local local_time = util.exec("date +\"%Y-%m-%d %H:%M:%S\""):trim()
	local uptime = tonumber(util.exec("cat /proc/uptime"):trim():split(" ")[1])
	local uptime_days = math.floor(uptime / 86400)
	local uptime_hours = math.floor((uptime % 86400) / 3600)
	local uptime_minutes = math.floor((uptime % 3600) / 60)
	local uptime_seconds = math.floor(uptime % 60)
	local uptime_str = string.format("%d days, %02d:%02d:%02d", uptime_days, uptime_hours, uptime_minutes, uptime_seconds)
	local load_average = util.exec("cat /proc/loadavg"):trim()
	print(hostname, machine, architecture, kernel_release, openwrt_release, local_time, uptime_str, load_average)
end

function pon_status()
	local _, _, ploam_status = string.find(util.exec("pon psg"):trim(), " current=(%d+) ")
	ploam_status = tonumber(ploam_status) / 10
	local cpu0_temp = (tonumber((fs.readfile("/sys/class/thermal/thermal_zone0/temp") or ""):trim()) or 0) / 1000
	local cpu1_temp = (tonumber((fs.readfile("/sys/class/thermal/thermal_zone1/temp") or ""):trim()) or 0) / 1000
	local eep50 = fs.readfile("/sys/class/pon_mbox/pon_mbox0/device/eeprom50", 256)
	local eep51 = fs.readfile("/sys/class/pon_mbox/pon_mbox0/device/eeprom51", 256)
	local optic_temp = eep51:byte(97) + eep51:byte(98) / 256
	local voltage = (bit.lshift(eep51:byte(99), 8) + eep51:byte(100)) / 10000
	local tx_bias = (bit.lshift(eep51:byte(101), 8) + eep51:byte(102)) / 500
	local tx_power = (bit.lshift(eep51:byte(103), 8) + eep51:byte(104)) / 10000
	local rx_power = (bit.lshift(eep51:byte(105), 8) + eep51:byte(106)) / 10000
	local eth_speed = tonumber((fs.readfile("/sys/class/net/eth0_0/speed") or ""):trim())
	local vendor_name = eep50:sub(21, 36):trim()
	local vendor_pn = eep50:sub(41, 56):trim()
	local vendor_rev = eep50:sub(57, 60):trim()
	local pon_mode = uci:get("gpon", "ponip", "pon_mode") or "xgspon"
	local module_type = util.exec(". /lib/8311.sh && get_8311_module_type"):trim() or "bfw"
	local active_bank = util.exec(". /lib/8311.sh && active_fwbank"):trim() or "A"
	print(ploam_status, cpu0_temp, cpu1_temp, optic_temp, voltage, tx_bias, tx_power, rx_power, eth_speed, vendor_name, vendor_pn, vendor_rev, pon_mode, module_type, active_bank)
end

function memory_status()
	local meminfo = util.exec("cat /proc/meminfo")
	local total_mem = tonumber(meminfo:match("MemTotal:%s+(%d+)"))
	local free_mem = tonumber(meminfo:match("MemFree:%s+(%d+)"))
	local buffers = tonumber(meminfo:match("Buffers:%s+(%d+)"))
	local cached = tonumber(meminfo:match("Cached:%s+(%d+)"))
	local used_mem = total_mem - free_mem - buffers - cached
	local used_mem_percent = used_mem / total_mem * 100
	print(total_mem, free_mem, buffers, cached, used_mem, used_mem_percent)
end

system_status()
pon_status()
memory_status()
