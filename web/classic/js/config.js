export class Config {
    constructor({baseUrl}) {
        this.baseUrl = baseUrl
    }
}

export function NewConfig({baseUrl}) {
    const config = new Config({baseUrl})
    return Object.freeze(config)
}
