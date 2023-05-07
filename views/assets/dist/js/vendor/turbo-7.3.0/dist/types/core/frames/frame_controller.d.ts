import { FrameElement, FrameElementDelegate, FrameLoadingStyle } from "../../elements/frame_element";
import { FetchRequest, FetchRequestDelegate } from "../../http/fetch_request";
import { FetchResponse } from "../../http/fetch_response";
import { AppearanceObserver, AppearanceObserverDelegate } from "../../observers/appearance_observer";
import { FormSubmission, FormSubmissionDelegate } from "../drive/form_submission";
import { Snapshot } from "../snapshot";
import { ViewDelegate, ViewRenderOptions } from "../view";
import { Locatable } from "../url";
import { FormSubmitObserver, FormSubmitObserverDelegate } from "../../observers/form_submit_observer";
import { FrameView } from "./frame_view";
import { LinkInterceptor, LinkInterceptorDelegate } from "./link_interceptor";
import { FormLinkClickObserver, FormLinkClickObserverDelegate } from "../../observers/form_link_click_observer";
import { VisitOptions } from "../drive/visit";
type VisitFallback = (location: Response | Locatable, options: Partial<VisitOptions>) => Promise<void>;
export type TurboFrameMissingEvent = CustomEvent<{
    response: Response;
    visit: VisitFallback;
}>;
export declare class FrameController implements AppearanceObserverDelegate<FrameElement>, FetchRequestDelegate, FormSubmitObserverDelegate, FormSubmissionDelegate, FrameElementDelegate, FormLinkClickObserverDelegate, LinkInterceptorDelegate, ViewDelegate<FrameElement, Snapshot<FrameElement>> {
    readonly element: FrameElement;
    readonly view: FrameView;
    readonly appearanceObserver: AppearanceObserver<FrameElement>;
    readonly formLinkClickObserver: FormLinkClickObserver;
    readonly linkInterceptor: LinkInterceptor;
    readonly formSubmitObserver: FormSubmitObserver;
    formSubmission?: FormSubmission;
    fetchResponseLoaded: (_fetchResponse: FetchResponse) => void;
    private currentFetchRequest;
    private resolveVisitPromise;
    private connected;
    private hasBeenLoaded;
    private ignoredAttributes;
    private action;
    readonly restorationIdentifier: string;
    private previousFrameElement?;
    private currentNavigationElement?;
    constructor(element: FrameElement);
    connect(): void;
    disconnect(): void;
    disabledChanged(): void;
    sourceURLChanged(): void;
    sourceURLReloaded(): Promise<void>;
    completeChanged(): void;
    loadingStyleChanged(): void;
    private loadSourceURL;
    loadResponse(fetchResponse: FetchResponse): Promise<void>;
    elementAppearedInViewport(element: FrameElement): void;
    willSubmitFormLinkToLocation(link: Element): boolean;
    submittedFormLinkToLocation(link: Element, _location: URL, form: HTMLFormElement): void;
    shouldInterceptLinkClick(element: Element, _location: string, _event: MouseEvent): boolean;
    linkClickIntercepted(element: Element, location: string): void;
    willSubmitForm(element: HTMLFormElement, submitter?: HTMLElement): boolean;
    formSubmitted(element: HTMLFormElement, submitter?: HTMLElement): void;
    prepareRequest(request: FetchRequest): void;
    requestStarted(_request: FetchRequest): void;
    requestPreventedHandlingResponse(_request: FetchRequest, _response: FetchResponse): void;
    requestSucceededWithResponse(request: FetchRequest, response: FetchResponse): Promise<void>;
    requestFailedWithResponse(request: FetchRequest, response: FetchResponse): Promise<void>;
    requestErrored(request: FetchRequest, error: Error): void;
    requestFinished(_request: FetchRequest): void;
    formSubmissionStarted({ formElement }: FormSubmission): void;
    formSubmissionSucceededWithResponse(formSubmission: FormSubmission, response: FetchResponse): void;
    formSubmissionFailedWithResponse(formSubmission: FormSubmission, fetchResponse: FetchResponse): void;
    formSubmissionErrored(formSubmission: FormSubmission, error: Error): void;
    formSubmissionFinished({ formElement }: FormSubmission): void;
    allowsImmediateRender({ element: newFrame }: Snapshot<FrameElement>, options: ViewRenderOptions<FrameElement>): boolean;
    viewRenderedSnapshot(_snapshot: Snapshot, _isPreview: boolean): void;
    preloadOnLoadLinksForView(element: Element): void;
    viewInvalidated(): void;
    willRenderFrame(currentElement: FrameElement, _newElement: FrameElement): void;
    visitCachedSnapshot: ({ element }: Snapshot) => void;
    private loadFrameResponse;
    private visit;
    private navigateFrame;
    proposeVisitIfNavigatedWithAction(frame: FrameElement, element: Element, submitter?: HTMLElement): void;
    changeHistory(): void;
    private handleUnvisitableFrameResponse;
    private willHandleFrameMissingFromResponse;
    private handleFrameMissingFromResponse;
    private throwFrameMissingError;
    private visitResponse;
    private findFrameElement;
    extractForeignFrameElement(container: ParentNode): Promise<FrameElement | null>;
    private formActionIsVisitable;
    private shouldInterceptNavigation;
    get id(): string;
    get enabled(): boolean;
    get sourceURL(): string | undefined;
    set sourceURL(sourceURL: string | undefined);
    get loadingStyle(): FrameLoadingStyle;
    get isLoading(): boolean;
    get complete(): boolean;
    set complete(value: boolean);
    get isActive(): boolean;
    get rootLocation(): URL;
    private isIgnoringChangesTo;
    private ignoringChangesToAttribute;
    private withCurrentNavigationElement;
}
export {};