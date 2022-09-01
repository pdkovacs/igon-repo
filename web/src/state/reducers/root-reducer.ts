import { combineReducers } from "redux";
import { NotificationSlice, notificationReducer } from "./notification-reducer";
import { appReducer, AppSlice } from "./app-reducer";
import { IconsSlice, iconsReducer } from "./icons-reducer";

export interface IconRepoState {
	readonly notifications: NotificationSlice;
	readonly app: AppSlice;
	readonly icons: IconsSlice;
}

export const rootReducer = combineReducers({
	notifications: notificationReducer,
	app: appReducer,
	icons: iconsReducer
});
