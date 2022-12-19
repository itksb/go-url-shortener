import {Config} from "./config.js"
import {ShortenApi} from "./api.js"
import {ShortenForm} from "./components/shorten-form.js";
import {BatchShortenForm} from "./components/batch-shorten-form.js";
import {ShortenHistory} from "./components/shorten-history.js";

class App {
    constructor({config, shortenApi}) {
        if (!config instanceof Config) throw new Error("wrong config type")
        this._config = config
        if (!shortenApi instanceof ShortenApi) throw new Error("wrong shorten Api type")
        this._shortenApi = shortenApi
    }

    run() {
        const q = document.querySelector.bind(document)
        this._shortenForm = new ShortenForm({
            querySelector: q,
            api: this._shortenApi
        })
        this._shortenForm.render()
        this._batchShortenForm = new BatchShortenForm({
            querySelector: q,
            api: this._shortenApi
        })
        this._batchShortenForm.render()
        this._shortenHistory = new ShortenHistory({
            querySelector: q,
            api: this._shortenApi
        })
        this._shortenHistory.render()
    }
}

export default App
