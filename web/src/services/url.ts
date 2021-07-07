const getUrl = window.location;
const pathName = getUrl.pathname.endsWith("/")
    ? getUrl.pathname.substring(0, getUrl.pathname.length - 1)
    : getUrl.pathname;
const baseUrl = getUrl.protocol + "//" + getUrl.host + pathName;

export default (path: string) => baseUrl + path;
