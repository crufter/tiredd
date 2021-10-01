# TIREDD

a toy Reddit clone built on free APIs

![tiredd](https://crufter.com/assets/images/tiredd.png)

[tiredd.org](https://tiredd.org)

## What is it

[This blog post](https://crufter.com/toy-open-source-reddit-clone) explains this project.

## How to run

Tiredd is built on free [Micro](https://m3o.com) APIs. To run, first get a Micro token, and save it into the environment variable MICRO_API_TOKEN.
Once that's done, you can simply run the backend with `go run main.go`. It will listen on port `8090`.

The frontend is an Angular application. Fork the repo and you can deploy for free on Netlify with the following settings:

```
Repo:               github.com/crufter/tiredd
Base dir:           tiredd
Build command       npm install && npm run build
Publish directory   tiredd/dist/tiredd
```

Change the repo to your fork.