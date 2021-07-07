import "normalize.css";
import "./global.scss";

import * as React from "react";
import * as ReactDOM from "react-dom";

import { IconList } from "./views/icon/icon-list";

ReactDOM.render(
    <IconList/>,
    document.getElementById("app")
);
