# HypernexCDN
The CDN (Content-Delivery Network) for Hypernex File Transfer

## Installing

Navigate to the [latest release](https://github.com/TigersUniverse/HypernexCDN/releases/latest) and download the target executable for your desired platform.

After installing simply run it once, and it will generate the config at `./config.toml`.

### Example Config

```toml
API_Server = 'http://hypernex.local/api/v1/'
AWS_key = 'access-key'
AWS_secret = 'access-secret'
AWS_endpoint = 'http://s3.local:9000'
AWS_region = 'us-1'
AWS_bucket = 'hypernex'
Mongo_URI = 'mongodb://mongodb-server'
REDIS_Address = 'redis://redis.local:6379/0'
PICS_Bucket = '/'
PUBLIC_PICS = 'root-pics-dir/'
```

Then run again and you should be able to see your CDN on `:3333`

## Proxy and Caching

Realistically, you can use any reverse proxy with HypernexCDN, but we will be working with Caddy for this example.

### Installing Caddy with cache-handler

To install Caddy with the [cache-handler](https://github.com/caddyserver/cache-handler) module, we'll first need to install [xcaddy](https://github.com/caddyserver/xcaddy) so we can easily build Caddy with the module.

Run the following commands to install xcaddy (Debian)

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/xcaddy/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-xcaddy-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/xcaddy/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-xcaddy.list
sudo apt update
sudo apt install xcaddy
```

After you install xcaddy, you can simply run xcaddy and provide the URLs for the cache-handler and your preferred [storage](https://github.com/darkweak/storages)

> [!WARNING]
>
> You **MUST** select a storage provider. For this example, we will be using `go-redis`, but you can select whichever one best suits your needs.

Simply run this command to build Caddy with the cache-handler and go-redis modules

```bash
xcaddy build --with github.com/caddyserver/cache-handler --with github.com/darkweak/storages/go-redis/caddy
```

After Caddy is built, it will be output to `./caddy`

### Configuring the Caddyfile

In the same directory as your caddy executable, create a new file called `Caddyfile`

At the top of the file, enter in your CDN information, then your normal Caddy configuration with the cache module referenced. Below is an example using go-redis installed locally.

```caddy
{
  cache {
    url 127.0.0.1:6379
  }
  regex {
    exclude /randomImage
    exclude /randomImage/*
  }
  ttl 1800s
}
examplecdn.yoururl.com {
  cache
  reverse_proxy :3333
}
```

### Running Caddy

To run caddy in the foreground, do

```bash
./caddy run
```
or to run in the background, do

```bash
./caddy start
```

Now that you run Caddy, you should be able to access your CDN with cache over SSL at the URL `https://examplecdn.yoururl.com/`!
