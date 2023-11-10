// Copyright 2022 Praetorian Security, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scan

// These import statements ensure that the init functions run in each plugin.
// When a new plugin is added, this list should be updated.

import (
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/dhcp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/dns"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/echo"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/ftp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/http"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/imap"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/ipmi"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/ipsec"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/jdwp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/kafka/kafkaNew"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/kafka/kafkaOld"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/ldap"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/linuxrpc"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/modbus"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/mqtt/mqtt3"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/mqtt/mqtt5"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/mssql"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/mysql"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/netbios"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/ntp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/openvpn"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/oracledb"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/pop3"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/postgresql"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/rdp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/redis"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/rsync"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/rtsp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/smb"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/smtp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/snmp"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/ssh"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/stun"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/telnet"
	_ "github.com/vela-ssoc/vela-radar/fingerprintx/plugins/services/vnc"
)
