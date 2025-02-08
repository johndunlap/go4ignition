# go4ignition
**_go4ignition_ is NOT a framework.** This must be distinctly understood or nothing wonderful can come from its use. By
design, there are minimal abstractions between its code and your code and there is no update mechanism. If you wish to
update, you must run `git pull` and work through the resulting merge conflicts. It is expected that you fork it, change
what you don't like, delete what you don't need, and add what's missing. Hopefully, this will help you get off the
ground faster, similar to a framework, without incurring the ongoing maintenance costs of a framework. This approach
isn't for everyone. Some people prefer using a framework and that's great. However, if you would prefer to avoid a
framework but struggle, as most of us do, with time constraints then _go4ignition_ might be exactly what you're looking
for.

# Shell Scripts
- **_reload.sh_**
This script has only been tested on Linux and may not work correctly on other operating systems. This is the core
of the developer experience. Running this script will compile your application, run it, and watch the project directory
for changes. When a file is modified it will kill your application, re-compile it, and restart it. A round trip
typically takes less than 2 seconds from the moment you save a file in your editor. This script will automatically
invoke `genpersistence.sh` if `sites/SITE/migrations.go` has been modified. The `genstatic.sh` script will be invoked
if a file within `sites/SITE/static` as modified. The `gentemplate.sh` script will be invoked if a file within
`sites/SITE/template` has been modified.

- **_add_site.sh_**
go4ignition projects support hosting multiple sites out of the box. The folder `sites/.skeleton` is used as a template
for newly created sites. For example, if you own the domain `example.com` you can run the command
`./add_site.sh example.com` followed by the command `./reload.sh` and your new site will be accessible at
[http://localhost.example.com:8002](http://localhost.example.com:8002). For this to work, you either need to add 
`127.0.0.1    localhost.example.com` to your `/etc/hosts` file OR define a DNS A record for `localhost.example.com`
which resolves to `127.0.0.1`.

- **_genimport.sh_**
This script generates `import.go`, which contains one import statement for each site that has been added
with `add_site.sh`. These import statement ensure that each site's `init()` method is called.

- **_genpersistence.sh_**
This script runs each site's `sites/SITE/migration.go` file against an empty database, introspects the empty database, and
generates persistence code for interacting with it.

- **_genstatic.sh_**
For each site, this script scans the `sites/SITE/static` directory for and generates `sites/SITE/static.go`, which
embeds all static files in the `sites/SITE/static` folder into the Go executable, along with metadata which allows them
to be downloaded from `StaticFileHandler`.

- **_gentemplate.sh_**
For each site, this script scans `sites/SITE/template` and generates `sites/SITE/template.go`, which embeds all
templates into the Go executable, along with metadata which allows them to be parsed and used by handlers. For all html
files outside the `sites/SITE/template/fragment` directory, a default handler is generated. The metadata in this file
allows these generated handlers to be automatically registered with the web server.

# Database
You can use whatever database you prefer but `SQLite` is the only database which is supported out of the box. By
default, databases are stored at `$HOME/.go4ignition/sites/SITE/SITE.db`. Sane default pragmas, which should afford
good SQLite performance, are provided in `sites/database.go` and are shared between sites. You may need to tune them
for your use case, but they should be good enough to get you off the ground. 

# Database Migrations
Each site contains a `sites/SITE/migrations.go` file. This file contains a string array which is, creatively, named
`migrations`. Each string in this array will be executed against the database in the order it appears in this array
during application startup. Initially, this array will only contain one entry which is the SQLite definition for the
`Migration` table. If you remove this, database migrations will not work because this table is used to track which
migrations have already been run.

# Port Number
The default port number is `8002` is set in `sites/registry.go`. Aside from code changes, this can be overridden with
the `--port` command line flag or the `G4I_PORT` environment variable.

# Hostname
You may be confused by the fact that all requests for `localhost:8002` fail. This is a
deliberate security precaution. There are always hackers are scanning IP addresses for open ports and when
they find open ports they probe those ports looking for vulnerabilities. If they're scanning addresses in known address
ranges, they may not even know your site exists. This is the circumstance this guards against. If your site is accessed
by IP address, IE: The HTTP Host header doesn't contain a known hostname, HTTP 403 is automatically returned. This
doesn't make your site secure by itself, but it does block a surprising amount of lazy malicious traffic. They're just
scanning addresses in known networks. The hostname of each site has a direct relationship with its folder name. If you
take the domain name and replace all the periods with underscores, you will get the name of the site folder. For
example, the domain `example.com` becomes `example_com`. So, if you call `./add_site.sh example.com` the folder
`sites/example_com` will be created. You do not need to recompile go4ignition applications differently for each
environment. It is strongly advised that you continue following this pattern. You're less likely to encounter bugs if
the only difference between a test deployment and a production deployment is where you copy the binary to. Toward that
end, whatever the domain name happens to be, local development can be done with `localhost.DOMAIN_NAME`. In our
previous example, we used `example.com`. This means we would do local development with the hostname
`localhost.example.com`.

# Reserved URL Space
| **URL**  | **Reason**                |
|----------|---------------------------|
| /api/    | Reserved for API calls    |
| /ws/     | Reserved for websockets   |
| /static/ | Reserved for static files |
