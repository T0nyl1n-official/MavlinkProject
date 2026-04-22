declare module 'sortablejs' {
    export interface SortableOptions {
        group?: string | {
            name: string;
            pull?: boolean | 'clone' | ((to: Sortable, from: Sortable, dragEl: HTMLElement, event: Event) => boolean);
            put?: boolean | ((to: Sortable, from: Sortable, dragEl: HTMLElement, event: Event) => boolean);
            revertClone?: boolean;
        };
        sort?: boolean;
        delay?: number;
        delayOnTouchOnly?: boolean;
        touchStartThreshold?: number;
        disabled?: boolean;
        store?: {
            get: (sortable: Sortable) => string[];
            set: (sortable: Sortable) => void;
        };
        handle?: string;
        draggable?: string;
        swapThreshold?: number;
        invertSwap?: boolean;
        invertedSwapThreshold?: number;
        direction?: 'horizontal' | 'vertical';
        forceFallback?: boolean;
        fallbackClass?: string;
        fallbackOnBody?: boolean;
        fallbackTolerance?: number;
        supportPointer?: boolean;
        emptyInsertThreshold?: number;
        removeCloneOnHide?: boolean;
        animation?: number;
        easing?: string;
        ghostClass?: string;
        chosenClass?: string;
        dragClass?: string;
        ignore?: string;
        filter?: string | ((event: Event, target: HTMLElement, sortable: Sortable) => boolean);
        preventOnFilter?: boolean;
        dataIdAttr?: string;
        swapClass?: string;
        hideSortableGhost?: boolean;
        multiDrag?: boolean;
        selectedClass?: string;
        removeCloneOnHide?: boolean;
        setData?: (dataTransfer: DataTransfer, dragEl: HTMLElement) => void;
        dropBubble?: boolean;
        dragoverBubble?: boolean;
        dataTransfer?: boolean;
        bubbleScroll?: boolean;
        scroll?: boolean | HTMLElement;
        scrollFn?: (offsetX: number, offsetY: number, originalEvent: Event) => void;
        scrollSensitivity?: number;
        scrollSpeed?: number;
        onStart?: (evt: SortableEvent) => void;
        onEnd?: (evt: SortableEvent) => void;
        onAdd?: (evt: SortableEvent) => void;
        onUpdate?: (evt: SortableEvent) => void;
        onSort?: (evt: SortableEvent) => void;
        onRemove?: (evt: SortableEvent) => void;
        onFilter?: (evt: SortableEvent) => void;
        onMove?: (evt: SortableMoveEvent, originalEvent: Event) => boolean | -1 | 1;
        onClone?: (evt: SortableEvent) => void;
        onChange?: (evt: SortableEvent) => void;
    }

    export interface SortableEvent {
        to: HTMLElement;
        from: HTMLElement;
        item: HTMLElement;
        clone?: HTMLElement;
        oldIndex: number | undefined;
        newIndex: number | undefined;
        oldDraggableIndex?: number;
        newDraggableIndex?: number;
        operation: string;
    }

    export interface SortableMoveEvent {
        related: HTMLElement;
        dragged: HTMLElement;
        draggedRect: DOMRect;
        relatedRect: DOMRect;
    }

    export class Sortable {
        constructor(element: HTMLElement, options?: SortableOptions);
        static create(element: HTMLElement, options?: SortableOptions): Sortable;
        option(name: string, value: any): void;
        option(name: string): any;
        destroy(): void;
        toArray(): string[];
        sort(order: string[]): void;
        save(): void;
        closest(el: HTMLElement, selector: string): HTMLElement | null;
        on(event: string, fn: (evt: any) => void): void;
        off(event: string, fn?: (evt: any) => void): void;
    }

    export default Sortable;
}