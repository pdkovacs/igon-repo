import { getType } from "typesafe-actions";
import { dismissMessage, MessagesAction, reportError, reportInfo } from "../actions/messages-actions";

interface Message {
	readonly error?: Error;
	readonly info?: string;
}

export interface MessagesSlice {
	readonly lastId: number;
	readonly idToMessageMap: {[msgId: string]: Message};
}

const initialErrors: MessagesSlice = {
	lastId: 0,
	idToMessageMap: {}
};

export const messagesReducer = (state: MessagesSlice = initialErrors, action: MessagesAction): MessagesSlice => {
	switch (action.type) {
		case getType(reportError): {
			return addError(action.payload, state);
		}
		case getType(reportInfo): {
			return addInfo(action.payload, state);
		}
		case getType(dismissMessage): {
			const messageId = action.payload;
			const nextMessageMap = Object.keys(state.idToMessageMap)
				.filter(id => id !== messageId)
				.reduce<{[id: string]: Message}>(
					(acc, curr) => {
						acc[curr] = state.idToMessageMap[curr];
						return acc;
					},
					{}
				);
			return {
				...state,
				idToMessageMap: nextMessageMap
			};
		}
		default: {
			return state;
		}
	}
};

const addError = (err: Error, messages: MessagesSlice): MessagesSlice => {
	const nextId = messages.lastId + 1;
	return {
		...messages,
		lastId: nextId,
		idToMessageMap: {
			...messages.idToMessageMap,
			[nextId.toString()]: {
				error: err
			}
		}
	};
};

const addInfo = (info: string, messages: MessagesSlice): MessagesSlice => {
	const nextId = messages.lastId + 1;
	return {
		...messages,
		lastId: nextId,
		idToMessageMap: {
			...messages.idToMessageMap,
			[nextId.toString()]: {
				info: info
			}
		}
	};
};
