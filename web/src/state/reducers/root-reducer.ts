import { combineReducers } from "redux";
import { NotificationSlice, notificationReducer } from "./notification-reducer";
import { appReducer, AppSlice } from "./app-reducer";

export interface IconRepoState {
	readonly notifications: NotificationSlice;
	readonly app: AppSlice;
}

export const rootReducer = combineReducers({
	notifications: notificationReducer,
	app: appReducer
});
