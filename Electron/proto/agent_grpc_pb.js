// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var agent_pb = require('./agent_pb.js');

function serialize_agent_CommandMessage(arg) {
  if (!(arg instanceof agent_pb.CommandMessage)) {
    throw new Error('Expected argument of type agent.CommandMessage');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_agent_CommandMessage(buffer_arg) {
  return agent_pb.CommandMessage.deserializeBinary(new Uint8Array(buffer_arg));
}


var AgentServiceService = exports.AgentServiceService = {
  streamMessage: {
    path: '/agent.AgentService/StreamMessage',
    requestStream: true,
    responseStream: true,
    requestType: agent_pb.CommandMessage,
    responseType: agent_pb.CommandMessage,
    requestSerialize: serialize_agent_CommandMessage,
    requestDeserialize: deserialize_agent_CommandMessage,
    responseSerialize: serialize_agent_CommandMessage,
    responseDeserialize: deserialize_agent_CommandMessage,
  },
};

exports.AgentServiceClient = grpc.makeGenericClientConstructor(AgentServiceService, 'AgentService');
