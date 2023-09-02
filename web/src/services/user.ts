import getEndpointUrl from "./url";
import { getData } from "./fetch-utils";

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

export const fetchUserInfo: () => Promise<UserInfo> = () => getData<void, UserInfo>("/user", 200)
.then(
    userInfo => ({
			...userInfo,
			authenticated: true
    })
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
