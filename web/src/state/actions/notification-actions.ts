import { ActionType, createAction } from "typesafe-actions";

export const receiveNotification = createAction("show-notification")<string>();
export const dismissNotification = createAction("dismiss-notification")<string>();

export type NotificationAction = (
    ActionType<typeof receiveNotification> |
    ActionType<typeof dismissNotification>
)
