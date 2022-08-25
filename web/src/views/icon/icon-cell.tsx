import * as React from "react";
import { IconDescriptor } from "../../services/icon";
import { preferredIconfileUrl } from "../../services/icon";

import "./icon-cell.styl";

export interface IconCellProps {
    readonly icon: IconDescriptor;
    readonly reqestDetails: () => void;
}

export const IconCell = (props: IconCellProps) =>
    <div className="icon-cell" onClick={props.reqestDetails}>
        <div className="icon-preview">
            <img src={preferredIconfileUrl(props.icon)} height="30"/>
        </div>
        <span className="icon-name">{props.icon.name}</span>
    </div>;
