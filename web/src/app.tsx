import React, { useEffect } from "react";
import { IconList } from "./views/icon/icon-list";
import { useDispatch, useSelector } from "react-redux";
import { fetchConfigAction, fetchUserInfoAction } from "./state/actions/app-actions";
import type {} from "redux-thunk/extend-redux";

import "./app.styl";
import { IconRepoState } from "./state/reducers/root-reducer";
import { LoginDialog } from "./views/login";
import { useReporters } from "./utils/use-reporters";

import "./services/notification";
import { useNotifications } from "./utils/use-notifications";

export const App = () => {

	const authenticated = useSelector((state: IconRepoState) => state.app.userInfo.authenticated);
	
	const dispatch = useDispatch();

	const { reportInfo } = useReporters();

	useEffect(() => {
		dispatch(fetchConfigAction());
		dispatch(fetchUserInfoAction());
	}, []);

	useEffect(() => {
		if (authenticated) {
			reportInfo("You are logged in (again)!");
		}
	}, [authenticated]);

	useNotifications();


	return <div className="iconrepo-app">
		<LoginDialog open={!authenticated} loginUrl="some URL"/>
		<IconList/>
	</div>;
};
