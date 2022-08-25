import { configureStore } from "@reduxjs/toolkit";
import { rootReducer } from "./reducers/root-reducer";


const initialState = {};

const store = configureStore({
	middleware: (getDefaultMiddleware) => getDefaultMiddleware(),
	reducer: rootReducer,
	preloadedState: initialState
});


export default store;
