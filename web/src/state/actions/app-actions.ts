import { ActionType, createAction } from "typesafe-actions";
import { Config, fetchConfig } from "../../services/config";
import { fetchUserInfo, UserInfo } from "../../services/user";
import { AppThunk } from "./base";

export const loginNeeded = createAction("app/prompt-for-login")<boolean>();

export const fetchConfigSuccess = createAction("app/fetch-config-success")<Config>();
export const fetchConfigFailure = createAction("app/fetch-config-failure")<Error>();
export const fetchConfigAction: () => AppThunk = ()  => {
	return dispatch => {
		return fetchConfig()
		.then(
			config => dispatch(fetchConfigSuccess(config)),
			error => dispatch(fetchConfigFailure(error))
		);
	};
};

export type ConfigAction = (
	ActionType<typeof fetchConfigAction> |
	ActionType<typeof fetchConfigSuccess> |
	ActionType<typeof fetchConfigFailure>
)

export const fetchUserInfoSuccess = createAction("app/fetch-userinfo-success")<UserInfo>();
export const fetchUserInfoFailure = createAction("app/fetch-userinfo-failure")<Error>();
export const fetchUserInfoAction: () => AppThunk = ()  => {
	return dispatch => {
		return fetchUserInfo()
		.then(
			userInfo => dispatch(fetchUserInfoSuccess(userInfo)),
			error => dispatch(fetchUserInfoFailure(error))
		);
	};
};

export type UserInfoAction = (
	ActionType<typeof loginNeeded> |
	ActionType<typeof fetchUserInfoAction> |
	ActionType<typeof fetchUserInfoSuccess> |
	ActionType<typeof fetchUserInfoFailure>
)
