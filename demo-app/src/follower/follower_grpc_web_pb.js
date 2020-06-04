/* eslint-disable */
/**
 * @fileoverview gRPC-Web generated client stub for follower
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var config_config_pb = require('../config/config_pb.js')
const proto = {};
proto.follower = require('./follower_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.follower.FollowerClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.follower.FollowerPromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.follower.PutRequest,
 *   !proto.follower.PutResponse>}
 */
const methodDescriptor_Follower_Put = new grpc.web.MethodDescriptor(
  '/follower.Follower/Put',
  grpc.web.MethodType.UNARY,
  proto.follower.PutRequest,
  proto.follower.PutResponse,
  /**
   * @param {!proto.follower.PutRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.PutResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.follower.PutRequest,
 *   !proto.follower.PutResponse>}
 */
const methodInfo_Follower_Put = new grpc.web.AbstractClientBase.MethodInfo(
  proto.follower.PutResponse,
  /**
   * @param {!proto.follower.PutRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.PutResponse.deserializeBinary
);


/**
 * @param {!proto.follower.PutRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.follower.PutResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.follower.PutResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.follower.FollowerClient.prototype.put =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/follower.Follower/Put',
      request,
      metadata || {},
      methodDescriptor_Follower_Put,
      callback);
};


/**
 * @param {!proto.follower.PutRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.follower.PutResponse>}
 *     A native promise that resolves to the response
 */
proto.follower.FollowerPromiseClient.prototype.put =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/follower.Follower/Put',
      request,
      metadata || {},
      methodDescriptor_Follower_Put);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.follower.GetRequest,
 *   !proto.follower.GetResponse>}
 */
const methodDescriptor_Follower_Get = new grpc.web.MethodDescriptor(
  '/follower.Follower/Get',
  grpc.web.MethodType.UNARY,
  proto.follower.GetRequest,
  proto.follower.GetResponse,
  /**
   * @param {!proto.follower.GetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.GetResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.follower.GetRequest,
 *   !proto.follower.GetResponse>}
 */
const methodInfo_Follower_Get = new grpc.web.AbstractClientBase.MethodInfo(
  proto.follower.GetResponse,
  /**
   * @param {!proto.follower.GetRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.GetResponse.deserializeBinary
);


/**
 * @param {!proto.follower.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.follower.GetResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.follower.GetResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.follower.FollowerClient.prototype.get =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/follower.Follower/Get',
      request,
      metadata || {},
      methodDescriptor_Follower_Get,
      callback);
};


/**
 * @param {!proto.follower.GetRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.follower.GetResponse>}
 *     A native promise that resolves to the response
 */
proto.follower.FollowerPromiseClient.prototype.get =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/follower.Follower/Get',
      request,
      metadata || {},
      methodDescriptor_Follower_Get);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.follower.SyncRequest,
 *   !proto.follower.SyncResponse>}
 */
const methodDescriptor_Follower_Sync = new grpc.web.MethodDescriptor(
  '/follower.Follower/Sync',
  grpc.web.MethodType.UNARY,
  proto.follower.SyncRequest,
  proto.follower.SyncResponse,
  /**
   * @param {!proto.follower.SyncRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.SyncResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.follower.SyncRequest,
 *   !proto.follower.SyncResponse>}
 */
const methodInfo_Follower_Sync = new grpc.web.AbstractClientBase.MethodInfo(
  proto.follower.SyncResponse,
  /**
   * @param {!proto.follower.SyncRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.SyncResponse.deserializeBinary
);


/**
 * @param {!proto.follower.SyncRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.follower.SyncResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.follower.SyncResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.follower.FollowerClient.prototype.sync =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/follower.Follower/Sync',
      request,
      metadata || {},
      methodDescriptor_Follower_Sync,
      callback);
};


/**
 * @param {!proto.follower.SyncRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.follower.SyncResponse>}
 *     A native promise that resolves to the response
 */
proto.follower.FollowerPromiseClient.prototype.sync =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/follower.Follower/Sync',
      request,
      metadata || {},
      methodDescriptor_Follower_Sync);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.follower.NotifyRequest,
 *   !proto.follower.NotifyResponse>}
 */
const methodDescriptor_Follower_Notify = new grpc.web.MethodDescriptor(
  '/follower.Follower/Notify',
  grpc.web.MethodType.UNARY,
  proto.follower.NotifyRequest,
  proto.follower.NotifyResponse,
  /**
   * @param {!proto.follower.NotifyRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.NotifyResponse.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.follower.NotifyRequest,
 *   !proto.follower.NotifyResponse>}
 */
const methodInfo_Follower_Notify = new grpc.web.AbstractClientBase.MethodInfo(
  proto.follower.NotifyResponse,
  /**
   * @param {!proto.follower.NotifyRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.follower.NotifyResponse.deserializeBinary
);


/**
 * @param {!proto.follower.NotifyRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.follower.NotifyResponse)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.follower.NotifyResponse>|undefined}
 *     The XHR Node Readable Stream
 */
proto.follower.FollowerClient.prototype.notify =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/follower.Follower/Notify',
      request,
      metadata || {},
      methodDescriptor_Follower_Notify,
      callback);
};


/**
 * @param {!proto.follower.NotifyRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.follower.NotifyResponse>}
 *     A native promise that resolves to the response
 */
proto.follower.FollowerPromiseClient.prototype.notify =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/follower.Follower/Notify',
      request,
      metadata || {},
      methodDescriptor_Follower_Notify);
};


module.exports = proto.follower;
