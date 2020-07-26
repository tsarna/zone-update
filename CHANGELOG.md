## Next (Unreleased)
 
 * If one is configured, the password file is now reloaded on SIGHUP.
 * It is no longer possible to use the single user/password flags or
   environment variables together with a password file.
 * Add unit tests for AtomicFile. Coverage now at 7.2% (Statements) 
 
## 0.2.0 (July 18, 2020)

FEATURES:

 * Add documentation in README.md and add a LICENSE.md
 * With --sequential-serial use a simple incrementing zone serial, instead of a date-based one.
 * Make the URL prefix for the API configurable
 * Optionally serve a /robots.txt that blocks robots.
 * Now supports HTTPS
 * Don't trust proxy headers (X-Real-IP/X-Forwarded-For) unless --trust-proxy is given
 * Replace separate --http-addr and -http-port wth --listen
 * Add a changelog.

## 0.1.1 (July 11, 2020)

BUG FIXES:
 * Build with CGO_ENABLED=0 to produce a static binary.

## 0.1.0 (July 10, 2020)

Initial release.
