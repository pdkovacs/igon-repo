import React, { useEffect } from "react";
import { IconList } from "./views/icon/icon-list";
import { useDispatch, useSelector } from "react-redux";
import { fetchConfigAction, fetchDeployConfigAction, fetchUserInfoAction } from "./state/actions/app-actions";
import type {} from "redux-thunk/extend-redux";
import getEndPointUrl from "./services/url";

import "./app.styl";
import { IconRepoState } from "./state/reducers/root-reducer";
import { LoginDialog } from "./views/login";

import "./services/notification";
import { useNotifications } from "./utils/use-notifications";

export const App = () => {

	const authenticated = useSelector((state: IconRepoState) => state.app.userInfo.authenticated);
	const backendBaseUrl = useSelector((state: IconRepoState) => state.app.backendAccess.baseUrl);

	const dispatch = useDispatch();

	useEffect(() => {
		dispatch(fetchDeployConfigAction());
	}, []);

	useEffect(() => {
		if (backendBaseUrl !== null) {
			dispatch(fetchUserInfoAction());

			if (authenticated) {
				dispatch(fetchConfigAction());
			}
		}
	}, [authenticated, backendBaseUrl]);

	useNotifications();

	return <div className="iconrepo-app">
		{ !authenticated && <LoginDialog open={!authenticated} loginUrl={getEndPointUrl("/login")}/> }
		{ authenticated && backendBaseUrl !== null && <IconList/> }
	</div>;
};
