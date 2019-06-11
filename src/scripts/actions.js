import {createActions} from 'redux-actions';

export const types = {
  ADD_EVENT_MESSAGE: 'ADD_EVENT_MESSAGE',
  UPDATE_TOPOLOGY: 'UPDATE_TOPOLOGY',
};

export const actions = createActions({
  [types.ADD_EVENT_MESSAGE]: (event)=>(event),
  [types.UPDATE_TOPOLOGY]: (topology)=>(topology),
});
