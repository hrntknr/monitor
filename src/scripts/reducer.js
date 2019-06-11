import {handleActions} from 'redux-actions';
import {types} from './actions';
import update from 'immutability-helper';

const initialState = {
  networkState: {},
  topology: {
    targets: {},
    connections: [],
  },
};

export default handleActions({
  [types.ADD_EVENT_MESSAGE]: (state, action)=>{
    const {channel, payload} = action.payload;
    switch (channel) {
    case 'poll_target': {
      return {
        ...state,
        networkState: payload,
      };
    }
    case 'trap_interface_state': {
      if (state.networkState[payload.id] == null) {
        return state;
      }
      return update(state, {
        networkState: {[payload.id]: {interfaces: {[payload.ifIndex]: {operStatus: {$set: payload.operStatus}}}}},
      });
    }
    default: {
      return state;
    }
    }
  },
  [types.UPDATE_TOPOLOGY]: (state, action)=>{
    return {
      ...state,
      topology: action.payload,
    };
  },
}, initialState);
