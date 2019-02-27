# A tiny realtime search tool for cloud accounts

`cloudsearch` allows you to index and search content on cloud services such as Google services 
(Gmail, Google Drive, etc) and Dropbox, directly from a command line.

Data gets indexed and stored on your device only - this doesn't utilize any intermediate service
for indexing. The only "intermediary" is an [auth gateway](http://github.com/herval/authgateway),
used for Oauth2 token exchanges, but you can always deploy your own auth, for extra independence.

## Usage

### Configure an account
> cloudsearch login <account type>

The available account types are `Dropbox` or `Google`.

In order for the OAuth2 loop to complete, `cloudsearch` will require your machine to accept inbound HTTP 
requests while adding an account. The default port is `65432`, but you can override it with the `--oauthPort` flag

### Search for content
> cloudsearch search foo

### Listing configured accounts
> cloudsearch accounts list

### Removing an account


# TO DO
- Fix result order (newer should be first?)
- Fix search w/ lowercase on titles (not working? eg email subject)
- Confluence
- Jira
- Network listener
- Thumbnails
- Sentry support? 
- Cache queries to only hit downstream after 1 minute if they successfully returned
- Fix empty query search on cache w/ type predicates
- "Basic" account setup

