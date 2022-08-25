import { getType } from "typesafe-actions";
import { AppInfo } from "../../services/config";
import { UserInfo } from "../../services/user";
import { fetchConfigSuccess, ConfigAction, UserInfoAction, fetchUserInfoSuccess, loginNeeded } from "../actions/app-actions";

export interface AppSlice {
	readonly appInfo: AppInfo;
	readonly userInfo: UserInfo;
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
	}
};

export const appReducer = (state: AppSlice = initialState, action: ConfigAction | UserInfoAction): AppSlice => {
	switch(action.type) {
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
				appInfo: action.payload
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
