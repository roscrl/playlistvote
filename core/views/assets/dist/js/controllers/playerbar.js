import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
    static targets = [
        "audio",
        "playOrPause",
        "currentlyPlayingTrackAlbumUri",
        "currentlyPlayingTrackAlbumImage",
        "currentlyPlayingTrack",
        "currentlyPlayingTrackName",
        "currentlyPlayingTrackArtists",
        "currentlyPlayingSpotifyPlaylistUri"
    ]

    connect() {
        if (!this.audioTarget) return

        if (!this.audioTarget.src) {
            const track = document.querySelector("[data-track]:not([disabled])")
            if (!track) return

            this.setState(track)
            this.element.classList.remove("hidden")
            this.element.classList.add("visible")
        }

        if (this.audioIsPlaying === undefined) this.audioIsPlaying = false

        this.boundHandleSpacebar = this.handleSpacebar.bind(this)
        document.addEventListener("keydown", this.boundHandleSpacebar)

        document.querySelectorAll("[data-track]").forEach(track => {
            track.addEventListener("click", this.play.bind(this))
        })

        document.addEventListener("turbo:before-cache", this.disconnect.bind(this))

        this.boundPause = this.pause.bind(this)
        window.addEventListener("beforeunload", this.boundPause)
    }

    /**
     * @param {HTMLElement} track
     */
    setState(track) {
        this.audioTarget.src = track.dataset.previewUrl
        this.currentlyPlayingTrackAlbumUriTarget.href = track.dataset.albumUri
        this.currentlyPlayingTrackAlbumImageTarget.src = track.dataset.albumImage
        this.currentlyPlayingTrackTarget.href = track.dataset.trackUri
        this.currentlyPlayingTrackTarget.innerText = track.querySelector("[data-name]").dataset.name

        const artists = track.querySelectorAll("[data-artist]")
        this.currentlyPlayingTrackArtistsTarget.innerHTML = ""
        artists.forEach(artist => {
            const artistElementCopy = artist.cloneNode(true)
            const innerArtistSpan = artistElementCopy.querySelector(".artist-name")

            innerArtistSpan.outerHTML = `<a href="${artist.dataset.artistUri}" class="${innerArtistSpan.classList.value} hover:underline">${artist.dataset.artist}</a>`

            this.currentlyPlayingTrackArtistsTarget.innerHTML += artistElementCopy.outerHTML
        })

        this.currentlyPlayingSpotifyPlaylistUriTarget.href = track.dataset.playlistUri
    }

    /**
     * @param {Event} event
     */
    handleSpacebar(event) {
        if (event.code !== "Space") return

        event.preventDefault()
        this.togglePlay()
    }

    togglePlay() {
        if (this.audioIsPlaying) {
            this.pause()
        } else {
            this.play()
        }
    }

    /**
     * @param {Event} event
     */
    play(event) {
        if (event) {
            const track = event.currentTarget
            this.setState(track)
        }

        this.audioTarget.play()
        this.setPauseIcon()
        this.audioIsPlaying = true
    }

    pause() {
        this.audioTarget.pause()
        this.setPlayIcon()
        this.audioIsPlaying = false
    }

    setPlayIcon() {
        this.playOrPauseTarget.innerHTML = `
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-8 h-8">
              <path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z"/>
            </svg>
       `
    }

    setPauseIcon() {
        this.playOrPauseTarget.innerHTML = `
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-8 h-8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5" />
          </svg>
        `
    }

    disconnect() {
        document.removeEventListener("keydown", this.boundHandleSpacebar)
        document.removeEventListener("turbo:before-cache", this.disconnect)
        window.removeEventListener("beforeunload", this.boundPause)
    }
}

