export class ShortenResponse {
    result = ""

    constructor({result}) {
        this.result = result;
    }
}

export class ShortenBatchItemResponse {
    correlation_id
    short_url

    constructor({correlation_id, short_url}) {
        this.correlation_id = correlation_id;
        this.short_url = short_url;
    }
}


export class ListUserURLResponseItem {
    ShortUrl = ""
    OriginalURL = ""

    constructor({short_url, original_url}) {
        this.ShortUrl = short_url;
        this.OriginalURL = original_url;
    }
}

class ListUserURLResponse extends Array {
    push(...items) {
        const b = items.every(value => value instanceof ListUserURLResponseItem)
        return super.push(...items);
    }
}


export class ShortenApi {
    constructor({baseUrl}, httpClient) {
        this._baseUrl = baseUrl;
        this._httpClient = httpClient;
        this._batchCount = 0;
    }

    /**
     * @param url
     * @returns {Promise<ShortenResponse>}
     */
    async shorten(url) {
        let response = await this._httpClient(`${this._baseUrl}/api/shorten`, {
            method: 'POST',
            credentials: 'include',
            body: JSON.stringify({url})
        })
        response = await response.json()
        return new ShortenResponse(response)
    }

    /**
     * @param url string[]
     * @returns {Promise<ShortenBatchItemResponse[]>}
     */
    async shortenBatch(urls) {
        let response = await this._httpClient(`${this._baseUrl}/api/shorten/batch`, {
            method: 'POST',
            credentials: 'include',
            body: JSON.stringify(
                [...urls]
                    .map((u) => {
                            return {correlation_id: `${this._batchCount++}`, original_url: u}
                        }
                    )
            )
        })
        response = await response.json()
        response = [...response]
        return response.map(v => new ShortenBatchItemResponse(v))
    }

    /**
     * @returns {Promise<ListUserURLResponse>}
     */
    async listUserURL() {
        const response = await this._httpClient(`${this._baseUrl}/api/user/urls`, {
            method: 'GET',
            credentials: 'include'
        });
        let result = await response.json()
        const resp = new ListUserURLResponse();
        [...result]
            .map(v => new ListUserURLResponseItem(v))
            .map(item => resp.push(item));
        return resp
    }
}
