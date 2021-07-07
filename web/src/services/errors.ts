export const throwError = (messageIntro: string, response: Response) => {
    return response.json().then(
        payloadError => {
            throw new Error(`${messageIntro}: ${payloadError.error}`);
        },
        error => { // reply is not a JSON
            throw new Error(`${messageIntro}: ${response.status} (${response.statusText})`);
        }
    );
};
