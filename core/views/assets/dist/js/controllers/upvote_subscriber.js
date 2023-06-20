import { Controller } from "@hotwired/stimulus"
import { connectStreamSource, disconnectStreamSource } from "@hotwired/turbo"
import { Routes } from "endpoints";

export default class extends Controller {
    eventSource

    async connect() {
        this.eventSource = new EventSource(Routes.PlaylistsUpvotesStream)

        connectStreamSource(this.eventSource)

        this.eventSource.addEventListener("connected", async (event) => {
            const playlistIds = []
            this.element.querySelectorAll("[data-playlist-id]").forEach((element) => {
                playlistIds.push(element.getAttribute("data-playlist-id"))
            })

            await fetch(Routes.PlaylistsUpvotesSubscribe, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({playlist_ids: playlistIds})
            })
        })
    }

    disconnect() {
        disconnectStreamSource(this.eventSource)
    }
}

