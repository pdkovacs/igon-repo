import { ActionType, createAction } from "typesafe-actions";

export const reportError = createAction("report-error")<Error>();
export const reportInfo = createAction("report-info")<string>();
export const dismissMessage = createAction("dismiss-message")<string>();

export type MessagesAction = (
	ActionType<typeof reportError> |
	ActionType<typeof reportInfo> |
	ActionType<typeof dismissMessage>
)
