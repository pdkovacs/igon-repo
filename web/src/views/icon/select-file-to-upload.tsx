import * as React from "react";

import "./select-file-to-upload.styl";
import { Icon } from "@blueprintjs/core";

interface SelectFileToUploadProps {
    readonly handleSelectedFile: (selectedFile: File) => void;
}

export const SelectFileToUpload = (props: SelectFileToUploadProps) =>
<div className="upload-container">
    <div className="upload--picture-card">
        <input type="file" name="" id=""
                onChange={event => {
                    props.handleSelectedFile(event.target.files[0]);
                }}
                accept="image/*"
                className="input-file"/>
        <Icon icon="plus"/>
    </div>
</div>;
