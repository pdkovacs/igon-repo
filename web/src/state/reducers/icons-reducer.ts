import { getType } from "typesafe-actions";
import { IconDescriptor } from "../../services/icon";
import { fetchIconsSuccess, IconsAction } from "../actions/icons-actions";

export interface IconsSlice {
	readonly allIcons: IconDescriptor[];
}

const initialState: IconsSlice = {
	allIcons: []
};

export const iconsReducer = (state: IconsSlice = initialState, action: IconsAction): IconsSlice => {
	switch(action.type) {
		case getType(fetchIconsSuccess): {
			return {
				...state,
				allIcons: action.payload
			};
		}
		default: {
			return state;
		}
	}
};
