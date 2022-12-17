import {NewConfig} from "./config.js";
import App from "./app.js"
import {ShortenApi} from "./api.js";


const config = NewConfig({baseUrl: "//localhost:8080"})
const shortenApi = new ShortenApi(config,  fetch.bind(window))
const app = new App({config, shortenApi})
app.run()
