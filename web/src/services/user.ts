import { List, Set } from "immutable";
import getEndpointUrl from "./url";
import { throwError } from "./errors";

export interface UserInfo {
    readonly username: string;
    readonly permissions: Set<string>;
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
    permissions: List()
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
        userInfo.permissions = Set(userInfo.permissions);
        userInfo.authenticated = true;
        return userInfo;
    }
);

export const logout = () => fetch(getEndpointUrl("/logout"), {
    method: "POST",
    mode: "no-cors",
    credentials: "include"
}).then(response => {
    window.location.assign(getEndpointUrl(""));
});

export const hasAddIconPrivilege = (user: UserInfo) => {
	console.log("-------- user: ", user)
    return user.permissions && user.permissions.has(privilegDictionary.CREATE_ICON);
}
export const hasUpdateIconPrivilege = (user: UserInfo) =>
    user.permissions && user.permissions.has(privilegDictionary.REMOVE_ICON);
