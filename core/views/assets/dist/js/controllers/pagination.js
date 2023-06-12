import { Controller } from "@hotwired/stimulus"
import { Routes } from "endpoints"

export default class extends Controller {
    static targets = ["loadingIcon"]
    isLoading = false

    async connect() {
        const observer = new IntersectionObserver(async entries => {
            for (const entry of entries) {
                if (entry.isIntersecting) {
                    if (this.isLoading) return
                    await this.more()
                }
            }
        }, { rootMargin: "400px" })

        observer.observe(this.element);
    }

    async more(_) {
        this.isLoading = true;
        this.showLoadingIcon();

        const lastPlaylistCard = document.querySelector(".playlist-card:last-child");
        const paginationId = lastPlaylistCard.getAttribute("data-pagination-id");

        const playlistsPaginationTopUrl = `${Routes.PlaylistsPaginationTop}${paginationId}`;

        const request = new Request(playlistsPaginationTopUrl, {
            method: "GET",
            headers: {
                "Accept": "text/vnd.turbo-stream.html"
            }
        });

        const response = await fetch(request)

        if (response.status === 204) {
            this.element.remove();
        } else if (response.status === 200) {
            const html = await response.text();
            Turbo.renderStreamMessage(html);
        }

        this.isLoading = false;
        this.hideLoadingIcon();
    }

    showLoadingIcon() {
        this.loadingIconTarget.classList.remove("hidden");
    }

    hideLoadingIcon() {
        this.loadingIconTarget.classList.add("hidden");
    }
}

