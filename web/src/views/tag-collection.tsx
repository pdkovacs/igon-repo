import { List, Set } from "immutable";
import * as React from "react";
import { Tag, ITagProps, Popover, MenuItem } from "@blueprintjs/core";

import "./tag-collection.styl";

export interface TagForSelection {
    readonly text: string;
    readonly toBeCreated: boolean;
}

export interface TagCollectionProps {
    readonly tags: List<string>;
    readonly selectedIndex?: number;
    readonly selectionChangeRequest?: (index: number) => void;
    readonly tagsAvailableForAddition?: Set<string>;
    readonly tagAdditionRequest?: (tagText: string) => void;
    readonly tagRemovalRequest?: (index: number) => void;
}

export interface TagCollectionState {
    readonly inputVisible: boolean;
    readonly inputValue: string;
}

export class TagCollection extends React.Component<TagCollectionProps, TagCollectionState> {

    constructor(props: TagCollectionProps) {
        super(props);
        this.state = {
            inputVisible: false,
            inputValue: ""
        };
    }

    public render() {
        return <div className="tag-collection">
            {this.props.tags.toArray().map((t, key) => this.createTag(t, key, this.props.selectedIndex === key))}
            {this.createInput()}
        </div>;
    }

    private createTag(tagText: string, index: number, selected: boolean) {
        return <Tag className="tag-collection-item"
                    minimal={true}
                    round={true}
                    key={index}
                    onClick={() => this.handleOnClick(index)}
                    onRemove={this.createOnRemoveHandler(index)}
                    data-selected={selected}>
                <span>{tagText}</span>
            </Tag>;
    }

    private createOnRemoveHandler(index: number) {
        return this.props.tagRemovalRequest
            ? (e: React.MouseEvent<HTMLButtonElement>, tagProps: ITagProps) => this.props.tagRemovalRequest(index)
            : undefined;
    }

    private handleOnClick(index: number) {
        if (this.props.selectionChangeRequest) {
            this.props.selectionChangeRequest(index);
        }
    }

    private createInput() {
        if (this.props.tagAdditionRequest) {
            return <Popover
                        isOpen={this.getItemsToAdd().length > 0}
                        minimal={true}>
                <input
                    value={this.state.inputValue}
                    onKeyPress={e => this.handleInputKey(e)}
                    onChange={e => this.handleInputChange(e)}/>
                {this.getSuggestionList()}
            </Popover>;
        } else {
            return null;
        }
    }

    private handleInputKey(e: React.KeyboardEvent<HTMLInputElement>) {
        if (e.charCode === 13) {
            this.requestAddition(this.state.inputValue);
        }
    }

    private handleInputChange(e: React.ChangeEvent<HTMLInputElement>) {
        this.setState({inputValue: e.target.value});
    }

    private getItemsToAdd() {
        const currentUserInput = this.state.inputValue;

        if (!this.props.tagsAvailableForAddition || !currentUserInput || !currentUserInput.length) {
            return [];
        }

        const tagsForSelection: List<TagForSelection> = this.props.tagsAvailableForAddition
                .subtract(this.props.tags)
                .filter(tagText => tagText.indexOf(currentUserInput) > -1)
                .map(t => ({
                    text: t,
                    toBeCreated: false
                }))
                .toList();

        return this.props.tags.indexOf(currentUserInput) === -1 &&
                    !this.props.tagsAvailableForAddition.contains(currentUserInput)
            ? tagsForSelection.push({
                    text: currentUserInput,
                    toBeCreated: true
                }).toArray()
            : tagsForSelection.toArray();
    }

    private requestAddition(tagToAdd: string) {
        this.props.tagAdditionRequest(tagToAdd);
        this.setState({inputValue: ""});
    }

    private getSuggestionList() {
        return <div className="suggestion-list">
                    {this.getItemsToAdd().map((item, index) =>
                        <MenuItem key={index}
                            text={item.toBeCreated ? `+ ${item.text} (New Tag)` : item.text}
                            onClick={() => this.requestAddition(item.text)}>
                        </MenuItem>)}
                </div>;

    }
}
