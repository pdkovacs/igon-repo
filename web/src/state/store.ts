import { configureStore } from "@reduxjs/toolkit";
import { rootReducer } from "./reducers/root-reducer";
import logger from "redux-logger";

const initialState = {};

const store = configureStore({
	middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(logger),
	reducer: rootReducer,
	preloadedState: initialState
});


export default store;
