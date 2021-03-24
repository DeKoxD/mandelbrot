// Based on OpenCL example from:
// https://subscription.packtpub.com/book/application_development/9781849692342/1/ch01lvl1sec12/an-example-of-opencl-program

#ifndef MANDELBROTOCL_H
#define MANDELBROTOCL_H

#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#define CL_TARGET_OPENCL_VERSION 220
#ifdef __APPLE__
#include <OpenCL/cl.h>
#else
#include <CL/cl.h>
#endif

const char *eval_coord_kernel =
"__kernel \n"
"void eval_coord_kernel(int resx, double lim_sq, int its, double dist, double ulx, double uly, __global bool *out) { \n"
"	int index = get_global_id(0); \n"
"	double x = (double)(index - (index/resx * resx)) * dist + ulx; \n"
"	double y = -(double)(index/resx) * dist + uly; \n"
"	double zx = 0, zy = 0, tzx; \n"
"	for(int i = 0; i < its; i++) { \n"
"		tzx = zx; \n"
"		zx = zx*zx - zy*zy + x; \n"
"		zy = tzx*zy + zy*tzx + y; \n"
"		if(zx*zx + zy*zy > lim_sq) { \n"
"			out[index] = false; \n"
"			return; \n"
"		} \n"
"	} \n"
"	out[index] = true; \n"
"}";

void compute_fractal(bool *out, double cx, double cy, double dist, int resx, int resy, double lim_sq, int its, size_t local_size) {
	const int out_size = resx*resy;

	// Get upper left corner x(real) and y(imaginary) coodinates
	const double ulx = cx + (-dist*((double)resx)/2 + dist/2);
	const double uly = cy + dist*((double)resy)/2 - dist/2;

  // Get platform and device information
  cl_platform_id * platforms = NULL;
  cl_uint     num_platforms;

  //Set up the Platform
  cl_int clStatus = clGetPlatformIDs(0, NULL, &num_platforms);
  platforms = (cl_platform_id *) malloc(sizeof(cl_platform_id)*num_platforms);
  clStatus = clGetPlatformIDs(num_platforms, platforms, NULL);

  //Get the devices list and choose the device you want to run on
  cl_device_id     *device_list = NULL;
  cl_uint           num_devices;

  clStatus = clGetDeviceIDs( platforms[0], CL_DEVICE_TYPE_GPU, 0,NULL, &num_devices);
  device_list = (cl_device_id *) malloc(sizeof(cl_device_id)*num_devices);
  clStatus = clGetDeviceIDs( platforms[0], CL_DEVICE_TYPE_GPU, num_devices, device_list, NULL);

  // Create one OpenCL context for each device in the platform
  cl_context context;
  context = clCreateContext( NULL, num_devices, device_list, NULL, NULL, &clStatus);

  // Create a command queue
  cl_command_queue command_queue = clCreateCommandQueueWithProperties(context, device_list[0], 0, &clStatus);

  // Create memory buffers on the device for each vector
  cl_mem Out_clmem = clCreateBuffer(context, CL_MEM_WRITE_ONLY, out_size * sizeof(bool), NULL, &clStatus);
	
	// // Copy the Buffer A and B to the device
  // clStatus = clEnqueueWriteBuffer(command_queue, A_clmem, CL_TRUE, 0, VECTOR_SIZE * sizeof(float), A, 0, NULL, NULL);
  // clStatus = clEnqueueWriteBuffer(command_queue, B_clmem, CL_TRUE, 0, VECTOR_SIZE * sizeof(float), B, 0, NULL, NULL);

  // Create a program from the kernel source
  cl_program program = clCreateProgramWithSource(context, 1,(const char **)&eval_coord_kernel, NULL, &clStatus);

  // Build the program
  clStatus = clBuildProgram(program, 1, device_list, NULL, NULL, NULL);
  
  // Create the OpenCL kernel
  cl_kernel kernel = clCreateKernel(program, "eval_coord_kernel", &clStatus);

  // Set the arguments of the kernel
  clStatus = clSetKernelArg(kernel, 0, sizeof(int), (void *)&resx);
  clStatus = clSetKernelArg(kernel, 1, sizeof(double), (void *)&lim_sq);
  clStatus = clSetKernelArg(kernel, 2, sizeof(int), (void *)&its);
  clStatus = clSetKernelArg(kernel, 3, sizeof(double), (void *)&dist);
  clStatus = clSetKernelArg(kernel, 4, sizeof(double), (void *)&ulx);
  clStatus = clSetKernelArg(kernel, 5, sizeof(double), (void *)&uly);
  clStatus = clSetKernelArg(kernel, 6, sizeof(cl_mem), (void *)&Out_clmem);

  // Execute the OpenCL kernel on the list
  // size_t local_size_ = 1;           // Process one item at a time
  // size_t global_size_ = out_size/local_size * local_size + local_size; // Process the entire lists
  const size_t global_size = resx*resy/local_size*local_size + local_size;
  clStatus = clEnqueueNDRangeKernel(command_queue, kernel, 1, NULL, &global_size, &local_size, 0, NULL, NULL);

  // Read the cl memory Out_clmem on device to the host variable out
  clStatus = clEnqueueReadBuffer(command_queue, Out_clmem, CL_TRUE, 0, out_size * sizeof(bool), out, 0, NULL, NULL);

  // Clean up and wait for all the comands to complete.
  clStatus = clFlush(command_queue);
  clStatus = clFinish(command_queue);

  // Finally release all OpenCL allocated objects and host buffers.
  clStatus = clReleaseKernel(kernel);
  clStatus = clReleaseProgram(program);
  clStatus = clReleaseMemObject(Out_clmem);
  clStatus = clReleaseCommandQueue(command_queue);
  clStatus = clReleaseContext(context);
  free(platforms);
  free(device_list);
}

#endif