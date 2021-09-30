# TIREDD

Tirred is an open source reddit clone

This Readme is copied from https://crufter.com/toy-open-source-reddit-clone.

As someone who likes the concept of Reddit but dislikes how the site is run, I keep launching short lived side projects.
I mostly use [Lemmy](https://github.com/LemmyNet/lemmy), but I can't justify running a SQL server and a VM for projects nobody cares about!

So as a toy project to test out [Micro](https://m3o.com) (disclaimer: I am a contributor), I decided to build a fun little Reddit inspired toy app: [Tiredd](tiredd.org) ([github repo](https://github.com/crufter/tiredd))

It is only about a thousand lines, and only a single page. The backend is a [single Go file](https://github.com/crufter/tiredd/blob/master/main.go), the frontend is a [single Angular component](https://github.com/crufter/tiredd/blob/master/tiredd/src/app/app.component.ts).

## What it has

### Reddit/HN-like ranking of posts

By a combination of score and time. Algorithm taken from here https://medium.com/hacking-and-gonzo/how-reddit-ranking-algorithms-work-ef111e33d0d9

### Anonymous posting and commenting

But only registered users can upvote/downvote to give the community ability to moderate.

### A very rudimentary way to moderate by admins

Make users mods by saving their user ids as a comma separated envar. Unlike normal users, they can vote on an item multiple times, and each time their vote counts as a random value between 4-17. I told you it's very rudimentary ; ).

### Hot/new page

Adjust to your taste by changing the minimum score on the frontend!

### Subs

Each post can be categorized under subs, just like Reddit.

## What it does not have

### Posts are not linkable

Is it a missing feature or a feature? It might actually prevent brigading :)).

## Comments are not tree like

Comments are ranked by score just like posts, but there is no reply, quote, or tree structure like on HN or Reddit.

### "Minor" features like logging out ; )

Things like logout or password change are missing, there is a chance I will add them!

## How to run

As previously stated, the main goal was to run this for free. There is no SQL server dependency, because it is using the [DB](https://m3o.com.db) and [User](https://m3o.com.user) services of Micro, all of which are free.

The go server is a single HTTP server which an be trivially translated to any FaaS provider (as it is stateless) to host the whole thing for free!

Enjoy!