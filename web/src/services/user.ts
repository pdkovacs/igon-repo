import getEndpointUrl from "./url";
import { throwError } from "./errors";

export interface UserInfo {
    readonly username: string;
    readonly permissions: string[];
    readonly authenticated: boolean;
}

const privilegDictionary = Object.freeze({
    CREATE_ICON: "CREATE_ICON",
    ADD_ICON_FILE: "ADD_ICON_FILE",
    REMOVE_ICON_FILE: "REMOVE_ICON_FILE",
    REMOVE_ICON: "REMOVE_ICON"
});

export const initialUserInfo = () => ({
    authenticated: false,
    username: "John Doe",
    permissions: [] as string[]
});

export const fetchUserInfo: () => Promise<UserInfo> = () => fetch(getEndpointUrl("/user"), {
    method: "GET",
    credentials: "include"
})
.then(response => {
    if (response.status !== 200) {
        return throwError("Failed to get user info", response);
    } else {
        return response.json();
    }
})
.then(
    userInfo => {
        userInfo.authenticated = true;
        return userInfo;
    }
);

export const logout = (idPlogoutUrl: string) => fetch(getEndpointUrl("/logout"), {
    method: "POST",
    mode: "no-cors",
    credentials: "include"
}).then(() => {
	window.location.assign(idPlogoutUrl);
});

export const hasAddIconPrivilege = (user: UserInfo) => {
	console.log("-------- user: ", user);
	return user.permissions && user.permissions.includes(privilegDictionary.CREATE_ICON);
};
export const hasUpdateIconPrivilege = (user: UserInfo) =>
    user.permissions && user.permissions.includes(privilegDictionary.REMOVE_ICON);
