import { createRef, Component, ComponentChildren } from "preact";
import { Modal } from "bootstrap";
import './bootstrap.scss';

interface ModalProps {
  children: ComponentChildren;
  buttons?: ComponentChildren;
  title: string;
  dismissHandler?: (hideModal?: () => void) => void;
}

export class MyModal extends Component<ModalProps, {}> {
  ref = createRef<HTMLDivElement>();

  constructor() {
    super();
  }

  componentDidMount() {
    Modal.getOrCreateInstance(this.ref.current!).show();
    this.ref.current?.addEventListener('hide.bs.modal', () => { this.props.dismissHandler && this.props.dismissHandler(); });
  }

  componentWillUnmount() {
    this.hide()
  }

  hide(): void {
    Modal.getOrCreateInstance(this.ref.current!).hide();
  }

  render() {
    return <div class="modal fade" ref={this.ref} tabIndex={-1} aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{this.props.title}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            {this.props.children}
          </div>
          <div class="modal-footer">
            {this.props.buttons}
          </div>
        </div>
      </div>
    </div>;
  }
}

interface ButtonProps {
  priority: "primary" | "secondary" | "danger"
  clickHandler?: () => void
  children: ComponentChildren
  disabled?: boolean;
}

export function Button(props:ButtonProps) {
  let classes: string = "btn btn-" + props.priority

  return <button type="button" disabled={props.disabled} class={classes} onClick={ () => { props.clickHandler ? props.clickHandler() : null; }}>{ props.children}</button>
}
