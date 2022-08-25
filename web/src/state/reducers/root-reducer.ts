import { combineReducers } from "redux";
import { messagesReducer as messagesReducer, MessagesSlice as MessagesSlice } from "./messages-reducer";
import { NotificationSlice, notificationReducer } from "./notification-reducer";
import { appReducer, AppSlice } from "./app-reducer";

export interface IconRepoState {
	readonly messages: MessagesSlice;
	readonly notifications: NotificationSlice;
	readonly app: AppSlice;
}

export const rootReducer = combineReducers({
	messages: messagesReducer,
	notifications: notificationReducer,
	app: appReducer
});
