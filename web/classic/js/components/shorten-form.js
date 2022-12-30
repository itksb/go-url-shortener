import debounce from "../libs/debounce.js";

const debounceDelay = 1000

export class ShortenForm {

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
        const $input = q('.single-input .shorten-form__input')
        const $ul = q('.single-input .shorten-response__list')
        /** @type {ShortenResponse}  */
        const shortenResponse = await this._api.shorten($input.value)
        const shortenUrl = shortenResponse.result
        const $li = q('#tmpl-shorten-response__list-item').content.cloneNode(true)
        const $a = $li.querySelector('a')
        $a.href = shortenUrl
        $a.text = shortenUrl
        $ul.appendChild($li)
    }

    render() {
        const q = this._querySelector
        const $btn = q('.single-input .shorten-form__button')
        $btn.addEventListener('click', debounce(this._onBtnClick.bind(this), debounceDelay))
    }
}
