function onRequest(event) {
    const r = event.request;
    if (r.headers.host.value !== "www.youtube.com") {
        return r;
    }

    const onlyVideoOnYoutube = "https://www.youtube.com/watch?v=dQw4w9WgXcQ";
    const referer = r.headers.referer;
    if (referer && referer.value === onlyVideoOnYoutube) {
        return r;
    }

    if (r.uri === "/watch" && r.querystring.v.value === "dQw4w9WgXcQ") {
        return r;
    }

    return {
        statusCode: 302,
        statusDescription: 'Found',
        headers: {
            location: { value: onlyVideoOnYoutube }
        }
    };
}
