import { ActionType, getType } from "typesafe-actions";
import { AppInfo } from "../../services/config";
import { UserInfo } from "../../services/user";
import { fetchConfigSuccess, ConfigAction, UserInfoAction, fetchUserInfoSuccess, loginNeeded, fetchDeployConfigSuccess } from "../actions/app-actions";

export interface AppSlice {
	readonly appInfo: AppInfo;
	readonly userInfo: UserInfo;
	readonly idPLogoutUrl: string;
	readonly backendUrl: string;
}

const initialState: AppSlice = {
	appInfo: {
		versionInfo: {
			version: "No data",
			commit: "No data"
		},
		appDescription: "No data"
	},
	userInfo: {
		permissions: [],
		username: "John Doe",
		authenticated: false
	},
	idPLogoutUrl: "/",
	backendUrl: null
};

export const appReducer = (state: AppSlice = initialState, action: ActionType<typeof fetchDeployConfigSuccess> | ConfigAction | UserInfoAction): AppSlice => {
	switch(action.type) {
		case getType(fetchDeployConfigSuccess): {
			return {
				...state,
				backendUrl: action.payload.backendUrl
			};
		}
		case getType(loginNeeded): {
			return {
				...state,
				userInfo: loginNeeded
					? {
							...state.userInfo,
							authenticated: false
						}
					: state.userInfo,
			};
		}
		case getType(fetchConfigSuccess): {
			return {
				...state,
				appInfo: action.payload.appInfo,
				idPLogoutUrl: action.payload.clientConfig.idPLogoutUrl
			};
		}
		case getType(fetchUserInfoSuccess): {
			return {
				...state,
				userInfo: action.payload
			};
		}
		default: {
			return state;
		}
	}
};
