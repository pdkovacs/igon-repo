import { List, Set } from "immutable";
import getEndpointUrl from "./url";
import { throwError } from "./errors";

export interface UserInfo {
    readonly username: string;
    readonly privileges: Set<string>;
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
    privileges: List()
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
        userInfo.privileges = Set(userInfo.privileges);
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

export const hasAddIconPrivilege = (user: UserInfo) =>
    user.privileges && user.privileges.has(privilegDictionary.CREATE_ICON);
export const hasUpdateIconPrivilege = (user: UserInfo) =>
    user.privileges && user.privileges.has(privilegDictionary.REMOVE_ICON);
