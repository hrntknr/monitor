import {connect} from 'react-redux';

function mapStateToProps(state) {
  return {
    ...state,
  };
}

function mapDispatchToProps(dispatch) {
  return {
  };
}

export default function(component) {
  component = connect(mapStateToProps, mapDispatchToProps)(component);
  return component;
}
