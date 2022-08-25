import React, { useEffect } from "react";
import { AppMessageList } from "./views/notifications/app-messages";
import { IconList } from "./views/icon/icon-list";
import { useDispatch, useSelector } from "react-redux";
import { fetchConfigAction, fetchUserInfoAction } from "./state/actions/app-actions";
import type {} from "redux-thunk/extend-redux";

import "./app.styl";
import { IconRepoState } from "./state/reducers/root-reducer";
import { LoginDialog } from "./views/login";
import { reportInfo } from "./state/actions/messages-actions";

export const App = () => {

	const authenticated = useSelector((state: IconRepoState) => state.app.userInfo.authenticated);

	const dispatch = useDispatch();

	useEffect(() => {
		dispatch(fetchConfigAction());
		dispatch(fetchUserInfoAction());
	}, []);

	useEffect(() => {
		if (authenticated) {
			dispatch(reportInfo("You are logged in (again)!"));
		}
	}, [authenticated]);

	return <div className="iconrepo-app">
		<AppMessageList/>
		<div className="main-screen">
			<LoginDialog open={!authenticated} loginUrl="some URL"/>
			<IconList/>
		</div>
	</div>;
};
