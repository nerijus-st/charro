[![Go Report Card](https://goreportcard.com/badge/github.com/nerijus-st/charro)](https://goreportcard.com/report/github.com/nerijus-st/charro)
# Charro
<p align="center">
  <img src="https://nerijust-st.s3.us-east-2.amazonaws.com/charro_gopher.png" width="150" alt="Charro">
</p>

Charro can generate your private Spotify playlist based on your most listened tracks from either Last.fm or Spotify services. It will also add your top track to <a href="https://open.spotify.com/playlist/3XN89Ie0dEP5cInfxN5S5j?si=NabA_j3qQpmG_av0iFA44w" target="_blank">public Charro playlist</a> if you chose to generate long term or overall playlist.

This was made purely for Go learning purposes and out of curiousity, but some might find this tool useful for generating Spotify playlists (me). There are plenty of things to improve and this is not end product by any means (I doubt it will ever be). Code reviews and/or suggestions are welcome.

I've included logging users to database just to check how many will try this. Errors too.

Made with Go, Postgres, Bootstrap. Hosted on Heroku, custom domain is proxied through Cloudflare to enforce SSL as Heroku won't allow certificates with free dynos.

Up and running at <a href="https://charro.me/">charro.me</a>

Too lazy to add tests, might do it later.
