import {createActions} from 'redux-actions';

export const types = {
  ADD_EVENT_MESSAGE: 'ADD_EVENT_MESSAGE',
};

export const actions = createActions({
  [types.ADD_EVENT_MESSAGE]: (event)=>(event),
});
