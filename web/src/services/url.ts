import store from "../state/store";

export default (path: string) => {
	const backendAccess = store.getState().app.backendAccess;
	const backendBaseUrl = backendAccess.baseUrl || "";
	const backendPathRoot = backendAccess.pathRoot || "";
	return backendBaseUrl + backendPathRoot + path;
};
