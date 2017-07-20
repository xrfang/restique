# RESTIQUE - A RESTful MySQL Query Proxy

RESTIQUE is a small utility providing easy MySQL access via HTTP interface. 

## Quick Start

### Building

Run `./build.go` under the root directory of this repository.  The build script
is written in Go and configured to be executed by `go run` which only works under
Linux.

The build script will carry out the following steps:

- Clone dependencies (which is defined in `depends`).
- Compile the project.

A standalone executable `bin/restique` will be generated if the build process
succeeds.

> **NOTE**: please ensure `build.go` has execution bit set and run it directly.
Do **not** use `go run build.go` as the script relies on the directory it is
being launched (`go run` will compile and launch it in a random temp directory).

### Running for the first time

1. Run `bin/restique -dsn-init` to generate a DSN configuration template, edit
the template, replace the placeholders with your database information. 
Specifically, you should:
    - replace `sample_conn` with a proper connection name,
	- enter correct [MySQL connection string (DSN)](https://github.com/go-sql-driver/mysql#dsn-data-source-name) 
	- and optionally edit the `memo` to describe the connection.
1. Run `bin/restique -user <username>` to add new users.  RESTIQUE will prompt
for password of the new user, as well as show a QR-code image for use by 2-Factor
authentication utilities, such as [FreeOTP](https://freeotp.github.io/).
1. Launch RESTIQUE: `bin/restique`
1. Open `http://localhost:32779/` in your browser.

### Further tweaks

1. You may read `conf/restique.conf.sample` for available options.  Note that
   all configuration options can also be specified via the command line (try
   `-help`).
1. If an option is specified both in the configuration file and on command line,
   the command line takes precedence.
1. There are a few command-line options not to be used in the configuration file:
    - `-dsn-init`: used to create a DSN configuration template
	- `-user` and `-pass`: used to create or update user authentication
1. RESTIQUE offers the capability to execute SQL query on multiple databases, which
   I call "parallel execution". To enable this feature, configure a special "combined"
   DSN in the configuration file, e.g.:

        "combined": {
			"driver": "[multi]",
			"dsn": "conn1,conn2",
			"memo": "this is a combined DSN"
		}
   In the above configuration, `[multi]` is the special name to identify a combined DSN,
   `conn1,conn2` are defined DSNs, which themselves cannot be combined.  Also note that
   although I call it "parallel", the query actually takes place sequentially, i.e.,
   conn1, then conn2...  If any error happens, RESTIQUE returns immediately, there is 
   NO transaction amongst multiple databases.

## APIs

RESTIQUE API design follows the [HATEOAS](https://en.wikipedia.org/wiki/HATEOAS)
principle. Simply visit `http://localhost:32779/api` to read about the APIs,
which should be self-explanatory.

## Security Considerations

### Transport Layer Security (TLS)

RESTIQUE provide multi-layer configurable security. The lowest layer is TLS.
If both `TLS_CERT` and `TLS_PKEY` are provided, RESTIQUE serves requests over
HTTPS.

### Two-Factor Authentication

By default, TOTP based two-factor authentication is enabled. To disable OTP,
you need to edit the authentication configuration (usually `restique_auth.json`)
and delete the `secret` for specific user.

Alternatively, you may consider delete the `pass` of a user to allow OTP code 
only login.  However, remove of both password and OTP authentication is not
allowed.

### Client IP Restriction

By default, RESTIQUE can be connected from any IP. It is possible to restrict
client IP addresses via the `CLIENT_CIDRS` directive. This is a CSV of
[CIDRs](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing) which the
client's IP address must match.

### Database Security

It is strongly recommended that the DSN configured with RESTIQUE is a read-only
connection to mitigate threats of improper usage, bugs or intrusion.

### Operation Logging

All operations are logged for the sake of auditing and problem solving. **WARNING**:
user name and password appears in the log in plain text form.  Also note that only
the first 5 lines of RESTIQUE replies are logged.

## Performance

RESTIQUE provides two limiting parameters: QUERY\_TIMEOUT and QUERY\_MAXROWS for
the `/query` API.  They are both 0 (disabled) by default.  To ensure data
integrity, `/exec` will not be limited by these parameters.

>**IMPORTANT**: RESTIQUE does **NOT** has the ability to analyze SQL statement.
In another word, `/query` and `/exec` has the same ability to execute any SQL
statement that is allowed by the DSN. The only difference is that `/query`
returns data (rows) while `/exec` returns "LastInsertId" and/or "RowsAffected".

## Constraints

### Database Support

RESTIQUE supports MySQL out-of-the-box. But it is very easy to add support for
other databases, such as [Postgres](https://github.com/lib/pq) or [SQLite](https://github.com/mattn/go-sqlite3).

### Platform Support

The building environment relies on Linux, as well as the password prompting
method, which used `stty` to suppress echo-ing. Other *NIX based systems may
also be supported, but I have no experience.

## Credits

* MySQL Driver: https://github.com/go-sql-driver/mysql
* OTP Support: https://github.com/pquerna/otp
* QR-Code: https://github.com/boombuler/barcode
* QR-Code: https://github.com/mdp/qrterminal
* BCrypt (password storage and validation): https://godoc.org/golang.org/x/crypto/bcrypt
* gopass by John Doak (johnsiilver at gmail dot com)
