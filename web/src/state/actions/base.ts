import { AnyAction } from "redux";
import { ThunkAction } from "redux-thunk";
import { IconRepoState } from "../reducers/root-reducer";


export type AppThunk<ReturnType = void> = ThunkAction<
  ReturnType,
  IconRepoState,
  unknown,
  AnyAction
>
