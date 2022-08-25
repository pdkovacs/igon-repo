import { getType } from "typesafe-actions";
import { dismissNotification, NotificationAction, receiveNotification } from "../actions/notification-actions";

export interface NotificationSlice {
	readonly lastId: number;
	readonly notificationId2NotificationMap: {[notificationId: string]: string};
}

const initialNotifications: NotificationSlice = {
	lastId: 0,
	notificationId2NotificationMap: {}
};

export const notificationReducer = (state: NotificationSlice = initialNotifications, action: NotificationAction): NotificationSlice => {
	switch (action.type) {
		case getType(receiveNotification): {
			return addNotification(action.payload, state);
		}
		case getType(dismissNotification): {
			const notificationId = action.payload;
			delete state.notificationId2NotificationMap[notificationId];
			return {
				...state,
				notificationId2NotificationMap: {
					...state.notificationId2NotificationMap
				}
			};
		}
		default: {
			return state;
		}
	}
};

const addNotification = (notification: string, notificationSlice: NotificationSlice): NotificationSlice => {
	const nextNotificationId = notificationSlice.lastId + 1;
	return {
		lastId: nextNotificationId,
		notificationId2NotificationMap: {
			...notificationSlice.notificationId2NotificationMap,
			[nextNotificationId.toString()]: notification
		}
	};
};
