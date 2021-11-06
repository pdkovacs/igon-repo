import "normalize.css";
import "./global.styl";

import * as React from "react";
import * as ReactDOM from "react-dom";

import { IconList } from "./views/icon/icon-list";

ReactDOM.render(
    <IconList/>,
    document.getElementById("app")
);
