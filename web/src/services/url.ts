import store from "../state/store";

export default (path: string) => (store.getState().app.backendUrl || "") + path;
