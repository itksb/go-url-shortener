export class ShortenHistory {

    _api;
    _querySelector;

    /**
     * @param querySelector {(string)=> Element }
     * @param api
     */
    constructor({querySelector, api}) {
        this._querySelector = querySelector;
        this._api = api;
    }

    render() {
        const q = this._querySelector
        const $ul = q('.shorten-history-list')
        this._api
            .listUserURL()
            .then(
                /**  @param listUserURLResponse  ListUserURLResponse */
                function (listUserURLResponse) {
                    listUserURLResponse.map(
                        /** @param listUserURLResponseItem ListUserURLResponseItem  */
                        function (listUserURLResponseItem) {
                            const $li = q('#tmpl-shorten-history-list__item').content.cloneNode(true)
                            const $aShort = $li.querySelector('.shorten-history-list__short-item')
                            const $aOrigin = $li.querySelector('.shorten-history-list__original-item')
                            $aShort.href = listUserURLResponseItem.ShortUrl
                            $aShort.text = listUserURLResponseItem.ShortUrl
                            $aOrigin.href = listUserURLResponseItem.OriginalURL
                            $aOrigin.text = listUserURLResponseItem.OriginalURL
                            $ul.appendChild($li)
                        })
                }
            )
    }
}
