import debounce from "../libs/debounce.js";
import {ShortenBatchItemResponse} from "../api.js";

const debounceDelay = 1000

export class BatchShortenForm {

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

    async _onBtnClick() {
        const q = this._querySelector
        const $input1 = q('.batch-input .shorten-form__input1')
        const $input2 = q('.batch-input .shorten-form__input2')
        const $ul = q('.batch-input .batch-shorten-response__list')
        /** @type {ShortenBatchItemResponse[]}  */
        const shortenBatchItemsResponse = await this._api.shortenBatch([$input1.value, $input2.value])

        for (const shortenItem of shortenBatchItemsResponse) {
            const $li = q('#tmpl-batch-shorten-response__list-item').content.cloneNode(true)
            const $a = $li.querySelector('a')
            $a.href = shortenItem.short_url
            $a.text = `short_url=${shortenItem.short_url}, correlation_id=${shortenItem.correlation_id}`
            $ul.appendChild($li)
        }

    }

    render() {
        const q = this._querySelector
        const $btn = q('.batch-input .shorten-form__button')
        $btn.addEventListener('click', debounce(this._onBtnClick.bind(this), debounceDelay))
    }
}
