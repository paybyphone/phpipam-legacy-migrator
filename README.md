# phpipam-legacy-migrator

This migrator takes a PHPIPAM MySQL database from a pre-1.0 version (namely
0.8), queries its VLANs, subnets, and IP addresses, and migrates them to a
recent PHPIPAM (1.2) via the API. The object here is to support a migration path
from a version of PHPIPAM that's too old to have a database upgrade path.

## What is Migrated

 * **VLANs**: Name, number (VLAN ID) and description are all migrated into the
   default L2 domain.
 * **Subnets**: Subnet CIDR (network and mask), description, and VLAN ID are all
   migrated to the section chosen by the user, or the default "Customers"
   section if not specified. Only IPv4 addresses are migrated. In addition to
   the above, the tool automatically detects parent subnets and add those
   subnets as master subnet IDs, meaning that old hierarchy is not preserved,
   however each subnet will cascade properly in the new DB (even if that was not
   the case before).
 * **Addresses**: IP address, description, and the hostname they belonged to are
   migrated. IPs are added to the subnets that were added in the previous
   step. Note that this tool does not migrate owner at this time.

## Installation

```
go get -u github.com/paybyphone/phpipam-legacy-migrator
```

This should install phpipam-legacy-migrator in your `GOPATH`'s bin directory.

## Preparing PHPIPAM

You probably want to start with a clean copy of PHPIPAM. Delete ALL data from it
(VLANs, subnets, and addresses specifically), enable the API, and configure an
application ID for use with it. Do not select crypt as a security type as the
tool does not support it.

This too

## Connecting to the DB

Connecting to the DB is pretty straightforward. We support both TCP and default
connections to MySQL. Using a non-default UNIX socket path is not supported (ie:
anything else but not specifying a hostname).

## Connecting to PHPIPAM

You can supply the options via the command line flags, or via the following
environment variables:

	* `PHPIPAM_APP_ID` for the application ID
	* `PHPIPAM_ENDPOINT_ADDR` for the API endpoint
	* `PHPIPAM_PASSWORD` for the PHPIPAM password
	* `PHPIPAM_USER_NAME` for the PHPIPAM username

## Command Line Options

```
Usage of phpipam-legacy-migrator:
  -appid string
    	The PHPIPAM application ID to use
  -dbhost string
    	The database host to connect to
  -dbname string
    	The name of the database to import data from (default "phpipam")
  -dbpassword string
    	The password for the database user
  -dbuser string
    	The database user to use (default "phpipam")
  -debug
    	Enable debug logging
  -endpoint string
    	The PHPIPAM endpoint to connect to
  -password string
    	The password for the PHPIPAM user
  -sectionid int
    	The section ID to add addresses to (default 1)
  -user string
    	The user to use when connecting to PHPIPAM
```

## License

```
Copyright 2017 PayByPhone Technologies, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
