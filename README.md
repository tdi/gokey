gokey
=====

Version: 0.2

[Keystok.com](http://keystok.com) client in Go language, both command line (`gokey`) and a library. 


Installation
============

`go get code.google.com/p/go.crypto/pbkdf2`
`go get github.com/tdi/gokey`


Configuration
=============

The access token can be either provided directly, `-a` in a CLI or as an argument in a library, 
or it can be set in a environmental variable `KEYSTOK_ACCESS_TOKEN`. Similarly, `KEYSTOK_CACHE_DIR`
can be set to point the location of the cache directory. 


CLI usage
=========

gokey [OPTIONS]

### -h 

Display a help message

### -a [access\_token]

Set access token string

### -c [cacheDir]

Set a cachedir location

### -nc 

Disable caching at all

### -v 

Set verbose mode on


If not set, caching is enabled with the default directory `~/.keystok`. It works in the
compatibility mode with the keystok Python client. 

Library usage
=============

```go 
import "keystok"

func main() {

  var access_token = "j9393939dj39dj2jd92jud92d2"

  var kc keystok.KeystokClient = keystok.GetKeystokClient(access_token)
  key:= kc.GetKey("somekey")

  var keys map[string]string
  keys = make(map[string]string)

  keys = kc.ListKeys()

}

```

You can also change Options of the `KeystokClient`:

```go

  kc.Opts.CacheDir = "some/other/location"
  kc.Opts.UseCache = false


```

AUTHOR
======

Copyright (c) Dariusz Dwornikowski




