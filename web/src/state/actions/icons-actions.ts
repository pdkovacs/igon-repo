import { ActionType, createAction } from "typesafe-actions";
import { describeAllIcons, IconDescriptor } from "../../services/icon";
import { AppThunk } from "./base";

export const fetchIconsSuccess = createAction("icons/fetch-icons-success")<IconDescriptor[]>();
export const fetchIconsFailure = createAction("icons/fetch-icons-failure")<Error>();
export const fetchIconsAction: () => AppThunk = ()  => {
	return dispatch => {
		return describeAllIcons()
		.then(
			icons => dispatch(fetchIconsSuccess(icons)),
			error => dispatch(fetchIconsFailure(error))
		);
	};
};

export type IconsAction = (
	ActionType<typeof fetchIconsAction> |
	ActionType<typeof fetchIconsSuccess> |
	ActionType<typeof fetchIconsFailure>
)
