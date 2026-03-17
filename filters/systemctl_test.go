package filters

import (
	"strings"
	"testing"
)

var systemctlStatusActiveFixture = `● bluetooth.service - Bluetooth service
     Loaded: loaded (/usr/lib/systemd/system/bluetooth.service; enabled; preset: enabled)
     Active: active (running) since Wed 2017-01-04 13:54:04 EST; 1 weeks 0 days ago
       Docs: man:bluetoothd(8)
   Main PID: 1234 (bluetoothd)
     Status: "Running"
      Tasks: 1 (limit: 4915)
     Memory: 2.1M
        CPU: 150ms
     CGroup: /system.slice/bluetooth.service
             └─1234 /usr/lib/bluetooth/bluetoothd

Jan 04 13:54:04 host systemd[1]: Starting Bluetooth service...
Jan 04 13:54:04 host bluetoothd[1234]: Bluetooth daemon 5.43
Jan 04 13:54:04 host systemd[1]: Started Bluetooth service.
Jan 04 13:54:05 host bluetoothd[1234]: Starting SDP server
Jan 04 13:54:05 host bluetoothd[1234]: Bluetooth management interface 1.14 initialized`

var systemctlStatusFailedFixture = `× nginx.service - A high performance web server and a reverse proxy server
     Loaded: loaded (/lib/systemd/system/nginx.service; enabled; vendor preset: enabled)
     Active: failed (Result: exit-code) since Tue 2024-03-12 10:00:00 UTC; 1min ago
    Process: 5678 ExecStart=/usr/sbin/nginx -g daemon on; master_process on; (code=exited, status=1/FAILURE)
   Main PID: 5678 (code=exited, status=1/FAILURE)
        CPU: 10ms

Mar 12 10:00:00 host systemd[1]: Starting A high performance web server...
Mar 12 10:00:00 host nginx[5678]: nginx: [emerg] bind() to 0.0.0.0:80 failed (98: Address already in use)
Mar 12 10:00:00 host systemd[1]: nginx.service: Main process exited, code=exited, status=1/FAILURE
Mar 12 10:00:00 host systemd[1]: nginx.service: Failed with result 'exit-code'.
Mar 12 10:00:00 host systemd[1]: Failed to start A high performance web server.`

func TestFilterSystemctlStatusActive(t *testing.T) {
	got, err := filterSystemctlStatus(systemctlStatusActiveFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "bluetooth.service - Bluetooth service") {
		t.Error("expected service name line")
	}
	if !strings.Contains(got, "Active: active (running)") {
		t.Error("expected active line")
	}
	if !strings.Contains(got, "Main PID: 1234 (bluetoothd)") {
		t.Error("expected PID line")
	}
	if !strings.Contains(got, "Tasks: 1 (limit: 4915), Memory: 2.1M, CPU: 150ms") {
		t.Errorf("expected combined stats, got: %q", got)
	}
	if strings.Contains(got, "CGroup:") {
		t.Error("CGroup should be stripped")
	}
	if strings.Contains(got, "Docs:") {
		t.Error("Docs should be stripped")
	}

	// Logs
	if !strings.Contains(got, "Bluetooth management interface 1.14 initialized") {
		t.Error("expected last log line")
	}
}

func TestFilterSystemctlStatusFailed(t *testing.T) {
	got, err := filterSystemctlStatus(systemctlStatusFailedFixture)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "nginx.service - A high performance web server") {
		t.Error("expected service name line")
	}
	if !strings.Contains(got, "Active: failed (Result: exit-code)") {
		t.Error("expected active line")
	}
	if !strings.Contains(got, "nginx: [emerg] bind() to 0.0.0.0:80 failed") {
		t.Error("expected error log line")
	}
}

func TestSystemctlRouted(t *testing.T) {
	f := get("systemctl", []string{"status", "nginx"})
	if f == nil {
		t.Fatal("expected filter for systemctl status, got nil")
	}

	f = get("systemctl", []string{"list-units"})
	if f == nil {
		t.Fatal("expected filter for systemctl list-units, got nil")
	}
}
