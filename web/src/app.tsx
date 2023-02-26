import React, { useEffect } from "react";
import { IconList } from "./views/icon/icon-list";
import { useDispatch, useSelector } from "react-redux";
import { fetchConfigAction, fetchDeployConfigAction, fetchUserInfoAction } from "./state/actions/app-actions";
import type {} from "redux-thunk/extend-redux";

import "./app.styl";
import { IconRepoState } from "./state/reducers/root-reducer";
import { LoginDialog } from "./views/login";
import { useReporters } from "./utils/use-reporters";

import "./services/notification";
import { useNotifications } from "./utils/use-notifications";

export const App = () => {

	const authenticated = useSelector((state: IconRepoState) => state.app.userInfo.authenticated);
	const backendUrl = useSelector((state: IconRepoState) => state.app.backendUrl);

	const dispatch = useDispatch();

	const { reportInfo } = useReporters();

	useEffect(() => {
		dispatch(fetchDeployConfigAction());
	}, []);

	useEffect(() => {
		if (backendUrl !== null) {
			dispatch(fetchUserInfoAction());

			if (authenticated) {
				reportInfo("You are logged in (again)!");
				dispatch(fetchConfigAction());
			}
		}
	}, [authenticated, backendUrl]);

	useNotifications();

	return <div className="iconrepo-app">
		{ !authenticated && <LoginDialog open={!authenticated} loginUrl="some URL"/> }
		{ authenticated && backendUrl !== null && <IconList/> }
	</div>;
};
