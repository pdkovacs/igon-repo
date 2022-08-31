import React, { useEffect } from "react";
import { IconList } from "./views/icon/icon-list";
import { useDispatch, useSelector } from "react-redux";
import { fetchConfigAction, fetchUserInfoAction } from "./state/actions/app-actions";
import type {} from "redux-thunk/extend-redux";

import "./app.styl";
import { IconRepoState } from "./state/reducers/root-reducer";
import { LoginDialog } from "./views/login";
import { useReporter } from "./services/app-messages";

export const App = () => {

	const authenticated = useSelector((state: IconRepoState) => state.app.userInfo.authenticated);

	const dispatch = useDispatch();

	const { reportInfo } = useReporter();

	useEffect(() => {
		dispatch(fetchConfigAction());
		dispatch(fetchUserInfoAction());
	}, []);

	useEffect(() => {
		if (authenticated) {
			reportInfo("You are logged in (again)!");
		}
	}, [authenticated]);

	return <div className="iconrepo-app">
		<LoginDialog open={!authenticated} loginUrl="some URL"/>
		<IconList/>
	</div>;
};
