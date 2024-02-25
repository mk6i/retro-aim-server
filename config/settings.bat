rem The port that the BOS service binds to.
set BOS_PORT=5191

rem The port that the auth service binds to.
set BUCP_PORT=5190

rem The port that the chat service binds to.
set CHAT_PORT=5192

rem The path to the SQLite database file. The file and DB schema are
rem auto-created if they doesn't exist.
set DB_PATH=oscar.sqlite

rem Disable password check and auto-create new users at login time. Useful for
rem quickly creating new accounts during development without having to register
rem new users via the management API.
set DISABLE_AUTH=true

rem Crash the server in case it encounters a client message type it doesn't
rem recognize. This makes failures obvious for debugging purposes.
set FAIL_FAST=false

rem Set logging granularity. Possible values: 'trace', 'debug', 'info', 'warn',
rem 'error'.
set LOG_LEVEL=info

rem The hostname that AIM clients connect to in order to reach OSCAR services
rem (BOS, BUCP, chat, etc). Make sure the hostname is reachable by all clients.
rem For local development, the default loopback address should work provided the
rem server and AIM client(s) are running on the same machine. For LAN-only
rem clients, a private IP address (e.g. 192.168..) or hostname should suffice.
rem For clients connecting over the Internet, specify your public IP address and
rem ensure that TCP ports 5190-5192 are open on your firewall.
set OSCAR_HOST=127.0.0.1

