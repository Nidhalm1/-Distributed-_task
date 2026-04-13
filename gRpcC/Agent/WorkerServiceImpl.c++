#include "../gRpc/file1.pb.h"
class WorkerServiceImpl final : public Worker::Service {
    grpc::Status Execute(grpc::ServerContext* context,
                         const TaskRequest* request,
                         TaskResponse* response) override {
        std::string cmd = request->command();
        for (const auto& arg : request->args()) {
            cmd += " " + arg;
        }
        std::cout << "Executing: " << cmd << std::endl;
        int ret = system(cmd.c_str());
        response->set_status(ret == 0 ? "OK" : "ERROR");
        response->set_output(std::to_string(ret));
        response->set_exit_code(ret);
        return grpc::Status::OK;
    }
}
{
private:
public:
    WorkerServiceImpl();
    ~WorkerServiceImpl();
};

WorkerServiceImpl::WorkerServiceImpl(/* args */)
{
}

WorkerServiceImpl::~WorkerServiceImpl()
{
}
